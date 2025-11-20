package eventhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EventoCulturaHandler maneja eventos de creación de eventos culturales
// Crea un tacho fake con id_tipo = 4 (Evento) para visualización en mapa
func EventoCulturaHandler(d amqp.Delivery) error {
	log.Println("Evento cultura.evento.crear recibido")
	log.Printf("Body raw completo: %s", string(d.Body))

	// PASO 1: Intentar parsear como objeto con campo "content"
	var wrapperWithContent struct {
		Content string `json:"content"`
	}

	var payloadMap map[string]interface{}

	if err := json.Unmarshal(d.Body, &wrapperWithContent); err == nil && wrapperWithContent.Content != "" {
		log.Println("Detectado formato con campo 'content' (JSON string)")

		// Parsear el content (que es un JSON string)
		var contentData map[string]interface{}
		if err := json.Unmarshal([]byte(wrapperWithContent.Content), &contentData); err != nil {
			log.Printf("Error parseando content string: %v", err)
			return fmt.Errorf("error parsing content: %v", err)
		}

		// Extraer el campo "data" que tiene el evento real
		if data, ok := contentData["data"].(map[string]interface{}); ok {
			payloadMap = data
		} else {
			log.Println("No se encontro campo 'data', usando contentData completo")
			payloadMap = contentData
		}
	} else {
		// Si no tiene campo "content", parsear directo
		log.Println("No se detecto campo 'content' - parseando body directo")
		if err := json.Unmarshal(d.Body, &payloadMap); err != nil {
			log.Printf("Error parseando payload directo: %v", err)
			return err
		}
	}

	// PASO 2: Extraer datos del evento
	nombreEvento := getStringValue(payloadMap, "name")
	descripcionEvento := getStringValue(payloadMap, "description")
	fechaEventoStr := getStringValue(payloadMap, "date")
	horaEvento := getStringValue(payloadMap, "time")

	log.Printf("Evento: %s", nombreEvento)
	log.Printf("Descripcion: %s", descripcionEvento)
	log.Printf("Fecha: %s", fechaEventoStr)
	log.Printf("Hora: %s", horaEvento)

	// PASO 3: Extraer coordenadas y datos del lugar
	var lat, lng float64
	var nombreLugar, direccion, categoria string

	if culturalPlace, ok := payloadMap["culturalPlaceId"].(map[string]interface{}); ok {
		nombreLugar = getStringValue(culturalPlace, "name")
		categoria = getStringValue(culturalPlace, "category")

		log.Printf("Lugar: %s", nombreLugar)
		log.Printf("Categoria: %s", categoria)

		if contact, ok := culturalPlace["contact"].(map[string]interface{}); ok {
			direccion = getStringValue(contact, "address")
			log.Printf("Direccion: %s", direccion)

			if coordinates, ok := contact["coordinates"].(map[string]interface{}); ok {
				if coords, ok := coordinates["coordinates"].([]interface{}); ok && len(coords) >= 2 {
					// GeoJSON usa [lng, lat]
					lng = getFloatValue(coords[0])
					lat = getFloatValue(coords[1])
					log.Printf("Coordenadas: lat=%v, lng=%v", lat, lng)
				}
			}
		}
	}

	// Validar que tenemos coordenadas
	if lat == 0 && lng == 0 {
		log.Println("No se pudieron extraer coordenadas validas del evento")
		return fmt.Errorf("coordenadas invalidas")
	}

	// PASO 4: Parsear fecha del evento
	fechaEvento, err := time.Parse(time.RFC3339, fechaEventoStr)
	if err != nil {
		log.Printf("Error parseando fecha, usando fecha actual: %v", err)
		fechaEvento = time.Now()
	}

	// PASO 5: Crear descripción completa para el tacho
	descripcionCompleta := fmt.Sprintf("Evento: %s | Lugar: %s | Fecha: %s %s",
		nombreEvento, nombreLugar, fechaEvento.Format("02/01/2006"), horaEvento)

	observaciones := fmt.Sprintf("Categoria: %s | %s | Direccion: %s",
		categoria, descripcionEvento, direccion)

	log.Printf("Descripcion tacho: %s", descripcionCompleta)
	log.Printf("Observaciones: %s", observaciones)

	// PASO 6: Crear tacho fake en MongoDB
	mongoCollection, err := config.GetMongoCollection("tachos")
	if err != nil {
		log.Printf("Error obteniendo coleccion MongoDB: %v", err)
		return fmt.Errorf("error getting mongo collection: %v", err)
	}

	// Obtener el último ID de MongoDB para generar uno nuevo
	ctx := context.Background()
	var lastTacho struct {
		ID int `bson:"id"`
	}
	opts := options.FindOne().SetSort(map[string]int{"id": -1})
	err = mongoCollection.FindOne(ctx, map[string]interface{}{}, opts).Decode(&lastTacho)

	mongoID := 1 // ID por defecto si no hay documentos
	if err == nil {
		mongoID = lastTacho.ID + 1
	}

	// Extraer el ID del evento cultural
	idEvento := getStringValue(payloadMap, "_id")
	if idEvento == "" {
		idEvento = getStringValue(payloadMap, "id")
	}

	// Insertar en MongoDB con el formato correcto
	tachoDoc := map[string]interface{}{
		"id":        mongoID,
		"lat":       lat,
		"lon":       lng,
		"id_evento": idEvento, // ID del evento cultural
	}

	_, err = mongoCollection.InsertOne(ctx, tachoDoc)
	if err != nil {
		log.Printf("Error insertando tacho en MongoDB: %v", err)
		return fmt.Errorf("error inserting tacho in MongoDB: %v", err)
	}

	log.Printf("Tacho insertado en MongoDB con ID: %d", mongoID)

	// PASO 7: Crear tacho fake en MySQL con id_tipo = 4 (Evento) + campos opcionales
	query := `
		INSERT INTO Tacho (
			id_tipo, 
			id_estado,
			id_mongo, 
			capacidad,
			descripcion,
			observaciones,
			fecha_instalacion
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result := config.DB.Exec(query,
		4,                   // id_tipo = 4 (Evento Cultural)
		1,                   // id_estado = 1 (activo)
		mongoID,             // id_mongo (ID de MongoDB)
		0.0,                 // capacidad = 0 (tacho fake)
		descripcionCompleta, // descripcion (nombre evento + lugar + fecha)
		observaciones,       // observaciones (categoría + descripción + dirección)
		fechaEvento,         // fecha_instalacion (fecha del evento)
	)

	if result.Error() != nil {
		log.Printf("Error insertando tacho en MySQL: %v", result.Error())
		// Intentar eliminar de MongoDB si falla MySQL
		mongoCollection.DeleteOne(ctx, map[string]interface{}{"id": mongoID})
		return fmt.Errorf("error inserting tacho in MySQL: %v", result.Error())
	}

	// PASO 8: Obtener el ID generado en MySQL
	var tachoID int64
	if result := config.DB.Raw("SELECT LAST_INSERT_ID()").Scan(&tachoID); result.Error() != nil {
		log.Printf("Advertencia: No se pudo obtener ID generado: %v", result.Error())
	} else {
		log.Printf("Tacho creado exitosamente - ID MySQL: %d", tachoID)
		log.Printf("Evento: %s", nombreEvento)
		log.Printf("Lugar: %s", nombreLugar)
		log.Printf("Fecha: %s", fechaEvento.Format("02/01/2006 15:04"))
	}

	log.Println("Evento cultural procesado exitosamente")
	return nil
}

// getStringValue extrae un valor string de un map de forma segura
func getStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getFloatValue convierte un valor a float64 de forma segura
func getFloatValue(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}
