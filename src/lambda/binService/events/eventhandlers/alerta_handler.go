package eventhandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events/schemas"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AlertaVecinalHandler maneja alertas de emergencias (pendientes)
// Las guarda como reclamos con tipo_origen = 'emergencia'
func AlertaVecinalHandler(d amqp.Delivery) error {
	return handleEmergenciaAlert(d, "EN_PROCESO", true) // true = publicar evento automáticamente
}

// AlertaResueltaHandler maneja alertas de emergencias resueltas
// Cambia el estado a ESPERA_INFO sin disparar evento (se dispara al cambio manual a RESUELTO)
func AlertaResueltaHandler(d amqp.Delivery) error {
	return handleEmergenciaAlert(d, "ESPERA_INFO", false) // false = no publicar evento
}

// ProcessAlertaVecinalDirect procesa una alerta vecinal directamente sin RabbitMQ
// Para ser llamado desde endpoints HTTP
func ProcessAlertaVecinalDirect(alerta schemas.AlertaVecinalEvent, estadoForzado string, publicarEvento bool) error {
	log.Println("========================================")
	log.Printf("🚨 Emergencia recibida directamente - Estado: %s", estadoForzado)
	log.Println("========================================")

	// Extraer datos del evento
	idEmergencia := alerta.ID
	tipoEmergencia := alerta.TipoEmergencia
	prioridad := alerta.Prioridad
	estado := alerta.Estado
	lat := alerta.Ubicacion.Lat
	lng := alerta.Ubicacion.Lon

	// Contexto por defecto
	contexto := "Reporte directo desde sistema"

	var comuna *int
	// El campo comuna no existe en AlertaVecinalEvent, lo dejamos nil

	log.Println("Datos de la emergencia:")
	log.Printf("  - ID Emergencia: %s", idEmergencia)
	log.Printf("  - Tipo: %s", tipoEmergencia)
	log.Printf("  - Prioridad: %s", prioridad)
	log.Printf("  - Estado: %s", estado)
	log.Printf("  - Coordenadas: lat=%v, lng=%v", lat, lng)

	return processEmergencia(idEmergencia, tipoEmergencia, prioridad, estado, contexto, lat, lng, comuna, estadoForzado, publicarEvento, "")
}

// handleEmergenciaAlert es el handler común para ambos tipos de alertas
func handleEmergenciaAlert(d amqp.Delivery, estadoForzado string, publicarEvento bool) error {
	log.Println("========================================")
	log.Printf("🚨 Evento emergencias.alerta recibido - Estado: %s", estadoForzado)
	log.Println("========================================")

	log.Printf("Body raw: %s", string(d.Body))

	// Parsear el payload directamente (viene sin envelope estándar)
	var payload struct {
		ID        string                 `json:"id"`
		Source    string                 `json:"source"`
		Timestamp string                 `json:"timestamp"`
		Topic     string                 `json:"topic"`
		Payload   map[string]interface{} `json:"payload"`
	}

	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("❌ [AlertaVecinal] Error parseando mensaje: %v", err)
		return fmt.Errorf("error parsing message: %v", err)
	}

	// Extraer datos del payload interno
	innerPayload := payload.Payload

	prioridad := getStringFromMap(innerPayload, "prioridad")
	estado := getStringFromMap(innerPayload, "estado")
	tipoEmergencia := getStringFromMap(innerPayload, "tipo_emergencia")
	lat := getFloatFromMap(innerPayload, "lat")
	lng := getFloatFromMap(innerPayload, "lng")
	contexto := getStringFromMap(innerPayload, "contexto")
	idReclamo := getStringFromMap(innerPayload, "id_reclamo")
	idEmergencia := getStringFromMap(innerPayload, "id_emergencia")

	var comuna *int
	if comunaVal, ok := innerPayload["comuna"].(float64); ok {
		comunaInt := int(comunaVal)
		comuna = &comunaInt
	}

	log.Println("Datos de la emergencia:")
	log.Printf("  - ID Emergencia: %s", idEmergencia)
	log.Printf("  - Tipo: %s", tipoEmergencia)
	log.Printf("  - Prioridad: %s", prioridad)
	log.Printf("  - Estado: %s", estado)
	log.Printf("  - Contexto: %s", contexto)
	log.Printf("  - Coordenadas: lat=%v, lng=%v", lat, lng)
	log.Printf("  - Comuna: %v", comuna)

	// Validar que tenemos los datos mínimos
	if idEmergencia == "" {
		log.Println("❌ [AlertaVecinal] ID de emergencia faltante")
		return fmt.Errorf("id_emergencia es requerido")
	}

	return processEmergencia(idEmergencia, tipoEmergencia, prioridad, estado, contexto, lat, lng, comuna, estadoForzado, publicarEvento, idReclamo)
}

// processEmergencia contiene la lógica común para procesar emergencias
func processEmergencia(idEmergencia, tipoEmergencia, prioridad, estado, contexto string, lat, lng float64, comuna *int, estadoReclamo string, publicarEvento bool, idReclamo string) error {
	// INVERTIR coordenadas para mantener consistencia con la BD
	latStr := fmt.Sprintf("%v", lng) // INVERTIDO
	lngStr := fmt.Sprintf("%v", lat) // INVERTIDO

	// Crear título y descripción para el reclamo
	titulo := fmt.Sprintf("Emergencia: %s", tipoEmergencia)
	descripcion := fmt.Sprintf("Alerta - Tipo: %s | Contexto: %s | Prioridad: %s | Estado: %s",
		tipoEmergencia, contexto, prioridad, estado)

	direccion := fmt.Sprintf("Lat: %v, Lng: %v", lat, lng)

	// PASO 1: Crear tacho fake en MongoDB
	mongoID, err := createFakeTachoMongo(idEmergencia, lat, lng)
	if err != nil {
		log.Printf("❌ [Emergencia] Error creando tacho fake en MongoDB: %v", err)
		// Continuar sin tacho fake
		mongoID = 0
	}

	// PASO 2: Crear tacho fake en MySQL (si se creó en MongoDB)
	var idTachoMySQL int
	if mongoID > 0 {
		idTachoMySQL, err = createFakeTachoMySQL(mongoID)
		if err != nil {
			log.Printf("❌ [Emergencia] Error creando tacho fake en MySQL: %v", err)
			// Continuar sin tacho en MySQL
			idTachoMySQL = 0
		} else {
			log.Printf("✅ [Emergencia] Tacho fake creado - MongoDB ID: %d, MySQL ID: %d", mongoID, idTachoMySQL)
		}
	}

	// Verificar si ya existe un reclamo con este id_emergencia (guardado en descripcion)
	var existingReclamo struct {
		IDReclamo    int    `gorm:"column:id_reclamo"`
		EstadoActual string `gorm:"column:estado"`
	}

	// Buscar por descripcion que contiene el id_emergencia
	checkQuery := `SELECT id_reclamo, estado FROM Reclamos WHERE tipo_origen = 'emergencia' AND descripcion LIKE ? LIMIT 1`
	searchPattern := fmt.Sprintf("%%ID: %s%%", idEmergencia)

	result := config.DB.Raw(checkQuery, searchPattern).Scan(&existingReclamo)

	if result.Error() == nil && existingReclamo.IDReclamo > 0 {
		// Ya existe el reclamo - ACTUALIZAR
		log.Printf("🔄 [AlertaVecinal] Reclamo existente encontrado - ID: %d, Estado actual: %s → Nuevo estado: %s",
			existingReclamo.IDReclamo, existingReclamo.EstadoActual, estadoReclamo)

		// Solo actualizar si el estado cambió
		if existingReclamo.EstadoActual != estadoReclamo {
			updateQuery := `UPDATE Reclamos SET estado = ?, descripcion = ? WHERE id_reclamo = ?`
			if result := config.DB.Exec(updateQuery, estadoReclamo, descripcion, existingReclamo.IDReclamo); result.Error() != nil {
				log.Printf("❌ [AlertaVecinal] Error actualizando estado del reclamo: %v", result.Error())
				return fmt.Errorf("error actualizando reclamo: %v", result.Error())
			}

			log.Printf("✅ [AlertaVecinal] Reclamo actualizado - ID: %d, Nuevo Estado: %s",
				existingReclamo.IDReclamo, estadoReclamo)

			// Publicar evento automáticamente solo si está habilitado
			if publicarEvento {
				go func() {
					if err := publishEstadoViaHTTP(existingReclamo.IDReclamo, estadoReclamo); err != nil {
						log.Printf("⚠️  [AlertaVecinal] Error publicando estado via HTTP: %v", err)
					} else {
						log.Printf("📤 [AlertaVecinal] Evento publicado a BI automáticamente")
					}
				}()
			} else {
				log.Printf("ℹ️  [AlertaVecinal] Publicación de evento omitida (estado: %s)", estadoReclamo)
			}
		} else {
			log.Printf("ℹ️  [AlertaVecinal] Estado sin cambios - ID: %d, Estado: %s",
				existingReclamo.IDReclamo, estadoReclamo)
		}

		log.Println("========================================")
		log.Println("✅ Emergencia actualizada exitosamente")
		log.Println("========================================")
		return nil
	}

	// No existe - CREAR NUEVO
	log.Println("➕ [AlertaVecinal] Reclamo no existe - creando nuevo...")

	// Agregar ID emergencia a la descripcion para poder buscarlo después
	descripcion = fmt.Sprintf("%s | ID: %s", descripcion, idEmergencia)

	// Insertar en MySQL como reclamo con tipo_origen = 'emergencia'
	query := `
		INSERT INTO Reclamos (
			id_persona,
			id_reclamo_externo,
			titulo,
			descripcion,
			estado,
			prioridad,
			direccion,
			lat,
			lng,
			comuna,
			fecha_creacion,
			tipo_origen
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), 'emergencia')
	`

	// Usar el id_emergencia como identificador
	var idReclamoExterno *int64
	if idReclamo != "plaza" && idReclamo != "" && idReclamo != idEmergencia {
		var id int64
		if _, err := fmt.Sscanf(idReclamo, "%d", &id); err == nil {
			idReclamoExterno = &id
		}
	}

	if result := config.DB.Exec(query,
		1,                // id_persona (1 = sistema)
		idReclamoExterno, // id_reclamo_externo
		titulo,           // titulo
		descripcion,      // descripcion
		estadoReclamo,    // estado (RESUELTO o EN_PROCESO)
		prioridad,        // prioridad
		direccion,        // direccion
		latStr,           // lat
		lngStr,           // lng
		comuna,           // comuna
	); result.Error() != nil {
		log.Printf("❌ [AlertaVecinal] Error insertando reclamo en MySQL: %v", result.Error())
		return fmt.Errorf("error insertando reclamo: %v", result.Error())
	}

	// Obtener el ID generado
	var reclamoID int64
	if result := config.DB.Raw("SELECT LAST_INSERT_ID()").Scan(&reclamoID); result.Error() != nil {
		log.Printf("⚠️  [AlertaVecinal] Advertencia: No se pudo obtener ID generado: %v", result.Error())
	} else {
		log.Printf("✅ [AlertaVecinal] Emergencia guardada como reclamo - ID MySQL: %d, Estado: %s",
			reclamoID, estadoReclamo)
	}

	log.Println("========================================")
	log.Println("✅ Emergencia procesada exitosamente")
	log.Println("========================================")

	return nil
}

// Helper functions
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getFloatFromMap(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

// publishEstadoViaHTTP publica el cambio de estado haciendo un HTTP call interno al endpoint REST
func publishEstadoViaHTTP(reclamoID int, estado string) error {
	// Obtener el puerto de la app desde variable de entorno
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	url := fmt.Sprintf("http://localhost:%s/reclamos/%d/estado", appPort, reclamoID)

	payload := map[string]string{
		"estado":     estado,
		"comentario": "Estado actualizado automáticamente desde módulo de Emergencias",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %v", err)
	}

	// Crear request HTTP
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Ejecutar request con timeout
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// createFakeTachoMongo crea un tacho fake en MongoDB para emergencias
func createFakeTachoMongo(idEmergencia string, lat, lng float64) (int, error) {
	mongoCollection, err := config.GetMongoCollection("bins")
	if err != nil {
		return 0, fmt.Errorf("error getting mongo collection: %v", err)
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

	// INVERTIR coordenadas para mantener consistencia con tachos existentes
	// MongoDB espera {lat: lng, lon: lat} invertido
	log.Printf("🗺️  [Emergencia] Guardando en MongoDB - lat: %v, lon: %v (invertido)", lng, lat)

	// Insertar en MongoDB con coordenadas INVERTIDAS
	tachoDoc := map[string]interface{}{
		"id":            mongoID,
		"lat":           lng,          // INVERTIDO
		"lon":           lat,          // INVERTIDO
		"id_emergencia": idEmergencia, // ID de la emergencia
	}

	_, err = mongoCollection.InsertOne(ctx, tachoDoc)
	if err != nil {
		return 0, fmt.Errorf("error inserting in MongoDB: %v", err)
	}

	log.Printf("✅ [Emergencia] Tacho fake creado en MongoDB - ID: %d", mongoID)
	return mongoID, nil
}

// createFakeTachoMySQL crea un tacho fake en MySQL vinculado al tacho de MongoDB
func createFakeTachoMySQL(mongoID int) (int, error) {
	if config.DB == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	// id_tipo = 7 para emergencias (ajusta según tu BD)
	// id_estado = 1 para activo
	query := `
		INSERT INTO Tacho (id_tipo, id_estado, id_mongo, capacidad) 
		VALUES (?, ?, ?, ?)
	`

	result := config.DB.Exec(query,
		7,       // id_tipo = 7 (emergencia) - ajusta según tu esquema
		1,       // id_estado = 1 (activo)
		mongoID, // ID del documento de MongoDB
		100.0,   // capacidad por defecto
	)

	if result.Error() != nil {
		return 0, fmt.Errorf("error inserting tacho: %v", result.Error())
	}

	// Obtener el ID generado
	var tachoID int64
	if result := config.DB.Raw("SELECT LAST_INSERT_ID()").Scan(&tachoID); result.Error() != nil {
		return 0, fmt.Errorf("error getting inserted ID: %v", result.Error())
	}

	log.Printf("✅ [Emergencia] Tacho fake creado en MySQL - ID: %d", tachoID)
	return int(tachoID), nil
}
