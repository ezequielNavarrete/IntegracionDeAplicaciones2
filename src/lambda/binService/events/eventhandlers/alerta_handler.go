package eventhandlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	amqp "github.com/rabbitmq/amqp091-go"
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

	// Crear título y descripción para el reclamo
	titulo := fmt.Sprintf("Emergencia: %s", tipoEmergencia)
	descripcion := fmt.Sprintf("Alerta - Tipo: %s | Contexto: %s | Prioridad: %s | Estado: %s",
		tipoEmergencia, contexto, prioridad, estado)

	direccion := fmt.Sprintf("Lat: %v, Lng: %v", lat, lng)

	// INVERTIR coordenadas para mantener consistencia con la BD
	log.Printf("Coordenadas originales - lat: %v, lng: %v", lat, lng)
	log.Printf("Guardando invertidas en MySQL - lat: %v, lng: %v", lng, lat)

	latStr := fmt.Sprintf("%v", lng) // INVERTIDO
	lngStr := fmt.Sprintf("%v", lat) // INVERTIDO

	// El estado viene determinado por el routing key (pendiente=EN_PROCESO, resuelto=RESUELTO)
	estadoReclamo := estadoForzado

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
