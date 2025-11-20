package eventhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events/schemas"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ReclamoResiduoHandler procesa eventos de reclamos de residuos creados
// Formato: envelope estándar con payload específico de reclamo
func ReclamoResiduoHandler(d amqp.Delivery) error {
	log.Printf("📥 [ReclamoResiduo] Recibido evento: %s", d.RoutingKey)

	// PRIMERO: Loguear el body COMPLETO sin parsear
	log.Printf("📄 [ReclamoResiduo] Body raw completo: %s", string(d.Body))
	log.Printf("📏 [ReclamoResiduo] Tamaño del body: %d bytes", len(d.Body))

	// 1. Parsear el envelope estándar
	var envelope schemas.EventEnvelope
	if err := json.Unmarshal(d.Body, &envelope); err != nil {
		log.Printf("❌ [ReclamoResiduo] Error parsing envelope: %v", err)
		log.Printf("📄 [ReclamoResiduo] Body que falló: %s", string(d.Body))
		return fmt.Errorf("error parsing envelope: %v", err)
	}

	log.Printf("📦 [ReclamoResiduo] Envelope - ID: %s, Source: %s, Topic: %s, Timestamp: %s",
		envelope.ID, envelope.Source, envelope.Topic, envelope.Timestamp)
	log.Printf("📦 [ReclamoResiduo] Payload type: %T, length: %d", envelope.Payload, len(envelope.Payload))

	// Verificar si el envelope está vacío (el otro equipo NO respeta el formato estándar)
	var payload schemas.ReclamoResiduoPayload
	if envelope.ID == "" && envelope.Source == "" && envelope.Topic == "" && len(envelope.Payload) == 0 {
		log.Println("⚠️  [ReclamoResiduo] Envelope VACÍO - el módulo NO está usando formato estándar")
		log.Println("📦 [ReclamoResiduo] Parseando el body directamente como payload (sin envelope)")

		// El body ES el payload directamente (sin envelope)
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			log.Printf("❌ [ReclamoResiduo] Error parseando body directo como payload: %v", err)
			return fmt.Errorf("error parsing direct payload: %v", err)
		}

		log.Printf("✅ [ReclamoResiduo] Payload parseado directamente (sin envelope)")
	} else {
		// Envelope válido - parsear payload desde envelope.Payload
		log.Println("✅ [ReclamoResiduo] Envelope válido - parseando payload desde envelope")
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			log.Printf("❌ [ReclamoResiduo] Error parsing payload: %v", err)
			log.Printf("📄 Payload raw: %s", string(envelope.Payload))
			return fmt.Errorf("error parsing payload: %v", err)
		}
	}

	log.Printf("📋 [ReclamoResiduo] Payload - ID: %d, Título: %s, Prioridad: %s",
		payload.IDReclamo, payload.Titulo, payload.Prioridad)

	// Continuar con el procesamiento del payload
	return processReclamoPayload(payload)
}

// processReclamoPayload procesa el payload del reclamo (independientemente de si vino con envelope o no)
func processReclamoPayload(payload schemas.ReclamoResiduoPayload) error {
	// 3. Convertir coordenadas string a float64
	lat, lng, err := parseCoordinates(payload.Lat, payload.Lng)
	if err != nil {
		log.Printf("⚠️  [ReclamoResiduo] Coordenadas inválidas, usando valores por defecto: %v", err)
		// Usar coordenadas por defecto (centro de la ciudad o 0,0)
		lat = 0.0
		lng = 0.0
	}

	// INVERTIR coordenadas para mantener consistencia con la BD
	log.Printf("Coordenadas originales - lat: %v, lng: %v", lat, lng)
	latInvertido := lng // INVERTIDO
	lngInvertido := lat // INVERTIDO
	log.Printf("Guardando invertidas en MySQL - lat: %v, lng: %v", latInvertido, lngInvertido)

	// 4. Validar que tengamos conexión a MySQL
	if config.DB == nil {
		return fmt.Errorf("database connection not available")
	}

	// 5. Preparar valores por defecto
	prioridad := payload.Prioridad
	if prioridad == "" {
		prioridad = "Media"
	}

	fecha := payload.Fecha
	if fecha.IsZero() {
		fecha = time.Now()
	}

	// 6. Insertar reclamo en MySQL con coordenadas INVERTIDAS
	query := `
		INSERT INTO Reclamos (id_persona, id_subcategoria, titulo, descripcion, prioridad, estado, direccion, lat, lng, fecha, id_reclamo_externo)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result := config.DB.Exec(query,
		1, // id_persona = 1 (no refiere a ninguna persona, FK eliminada)
		payload.IDSubcategoria,
		payload.Titulo,
		payload.Descripcion,
		prioridad,
		"Pendiente", // Estado inicial
		payload.Direccion,
		latInvertido, // INVERTIDO
		lngInvertido, // INVERTIDO
		fecha,
		payload.IDReclamo, // id_reclamo_externo - ID del reclamo en el sistema de Reclamos
	)

	if result.Error() != nil {
		log.Printf("❌ [ReclamoResiduo] Error guardando reclamo en MySQL: %v", result.Error())
		return fmt.Errorf("error guardando reclamo: %v", result.Error())
	}

	// 7. Obtener el ID generado
	var reclamoID int64
	if result := config.DB.Raw("SELECT LAST_INSERT_ID()").Scan(&reclamoID); result.Error() != nil {
		log.Printf("⚠️  [ReclamoResiduo] Advertencia: No se pudo obtener ID generado: %v", result.Error())
	} else {
		log.Printf("✅ [ReclamoResiduo] Reclamo guardado exitosamente - ID MySQL: %d, ID Original: %d",
			reclamoID, payload.IDReclamo)
	}

	// 8. Crear tacho "fake" para visualización en el front
	tachoFakeID, err := createFakeTachoForReclamo(payload, lat, lng)
	if err != nil {
		log.Printf("⚠️  [ReclamoResiduo] Error creando tacho fake: %v (el reclamo se guardó correctamente)", err)
	} else {
		log.Printf("🗑️  [ReclamoResiduo] Tacho fake creado - ID: %d", tachoFakeID)
	}

	return nil
}

// createFakeTachoForReclamo crea un tacho "fake" en MongoDB y MySQL para representar visualmente el reclamo
func createFakeTachoForReclamo(payload schemas.ReclamoResiduoPayload, lat, lng float64) (int, error) {
	if config.DB == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	// 1. Determinar neighborhood basado en coordenadas
	// TODO: Implementar lógica de geolocalización para determinar barrio
	// Por ahora usar un valor por defecto o extraerlo del reclamo
	neighborhood := 1 // Valor por defecto

	// 2. Crear documento en MongoDB
	mongoCollection, err := config.GetMongoCollection("bins")
	if err != nil {
		return 0, fmt.Errorf("error getting mongo collection: %v", err)
	}

	// Obtener el último ID de MongoDB para generar uno nuevo
	ctx := context.Background()
	var lastBin struct {
		ID int `bson:"id"`
	}
	opts := options.FindOne().SetSort(map[string]int{"id": -1})
	err = mongoCollection.FindOne(ctx, map[string]interface{}{}, opts).Decode(&lastBin)

	mongoID := 1 // ID por defecto si no hay documentos
	if err == nil {
		mongoID = lastBin.ID + 1
	}

	// Insertar en MongoDB
	binDoc := map[string]interface{}{
		"id":           mongoID,
		"lat":          lng, // NOTA: En MongoDB tienes lon primero
		"lon":          lat,
		"neighborhood": neighborhood,
	}

	_, err = mongoCollection.InsertOne(ctx, binDoc)
	if err != nil {
		return 0, fmt.Errorf("error inserting fake tacho in MongoDB: %v", err)
	}

	log.Printf("✅ [ReclamoResiduo] Tacho fake creado en MongoDB - ID: %d", mongoID)

	// 3. Crear registro en MySQL
	query := `
		INSERT INTO Tacho (id_tipo, id_estado, id_mongo, capacidad) 
		VALUES (?, ?, ?, ?)
	`

	result := config.DB.Exec(query,
		6,       // id_tipo = 6 (Reclamo)
		1,       // id_estado = 1 (activo por defecto)
		mongoID, // ID del documento de MongoDB
		0.0,     // capacidad = 0 (tacho fake)
	)

	if result.Error() != nil {
		// Si falla MySQL, intentar eliminar de MongoDB
		mongoCollection.DeleteOne(ctx, map[string]interface{}{"id": mongoID})
		return 0, fmt.Errorf("error inserting fake tacho in MySQL: %v", result.Error())
	}

	// Obtener el ID generado en MySQL
	var tachoID int64
	if result := config.DB.Raw("SELECT LAST_INSERT_ID()").Scan(&tachoID); result.Error() != nil {
		return 0, fmt.Errorf("error getting inserted ID: %v", result.Error())
	}

	log.Printf("✅ [ReclamoResiduo] Tacho fake creado en MySQL - ID: %d", tachoID)

	// 4. Asignar características por defecto (todas en estado "Nulo" - prioridad 0)
	queryCaracteristicas := `
		INSERT INTO Lista_caracteristica_tacho (id_tacho, id_caracteristica)
		SELECT ?, ct.id_caracteristica
		FROM Caracteristica_tacho ct
		INNER JOIN Estado_caracteristica ec ON ct.id_estado_caracteristica = ec.id_estado_caracteristica
		WHERE ec.prioridad = 0
	`

	if result := config.DB.Exec(queryCaracteristicas, tachoID); result.Error() != nil {
		log.Printf("⚠️  [ReclamoResiduo] Advertencia: No se pudieron asignar características por defecto: %v", result.Error())
	}

	return int(tachoID), nil
}

// parseCoordinates convierte strings de coordenadas a float64
// Maneja casos de strings vacíos "" o valores inválidos
func parseCoordinates(latStr, lngStr string) (float64, float64, error) {
	if latStr == "" || lngStr == "" {
		return 0, 0, fmt.Errorf("coordenadas vacías")
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("latitud inválida: %v", err)
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("longitud inválida: %v", err)
	}

	return lat, lng, nil
}
