package eventhandlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events/schemas"
	amqp "github.com/rabbitmq/amqp091-go"
)

// EventoCulturaHandler maneja eventos de creación de eventos culturales
// Por ahora solo loguea el evento completo para entender la estructura del payload
// TODO: Una vez conocida la estructura, implementar creación de tacho fake con id_tipo = 4 (Evento)
func EventoCulturaHandler(d amqp.Delivery) error {
	log.Println("========================================")
	log.Println("📢 Evento cultura.evento.crear recibido")
	log.Println("========================================")

	// Loguear el body completo primero
	log.Printf("📄 Body raw: %s", string(d.Body))

	// Intentar parsear como envelope estándar primero
	var envelope schemas.EventEnvelope
	if err := json.Unmarshal(d.Body, &envelope); err != nil {
		log.Printf("❌ [EventoCultura] Error parseando envelope: %v", err)
		return fmt.Errorf("error parsing envelope: %v", err)
	}

	// Verificar si el envelope tiene datos válidos
	if envelope.ID == "" && envelope.Source == "" && envelope.Topic == "" {
		log.Println("⚠️  [EventoCultura] Envelope vacío - cultura NO está usando formato estándar")
		log.Println("📦 El payload está directamente en el body (sin envelope)")
		
		// El body ES el payload directamente
		var payloadMap map[string]interface{}
		if err := json.Unmarshal(d.Body, &payloadMap); err != nil {
			log.Printf("❌ [EventoCultura] Error parseando payload directo: %v", err)
			return err
		}

		// Loguear cada campo del payload
		log.Println("📋 Campos del payload (sin envelope):")
		for key, value := range payloadMap {
			log.Printf("  - %s: %v (tipo: %T)", key, value, value)
		}
	} else {
		// Envelope válido - procesar normalmente
		log.Printf("📦 Envelope ID: %s", envelope.ID)
		log.Printf("📦 Timestamp: %s", envelope.Timestamp)
		log.Printf("📦 Source: %s", envelope.Source)
		log.Printf("📦 Topic: %s", envelope.Topic)

		// Loguear el payload completo como string
		payloadBytes, err := json.Marshal(envelope.Payload)
		if err != nil {
			log.Printf("❌ Error serializando payload: %v", err)
			return err
		}
		log.Printf("📦 Payload completo (JSON): %s", string(payloadBytes))

		// Intentar parsear el payload como map para ver todos los campos
		var payloadMap map[string]interface{}
		if err := json.Unmarshal(envelope.Payload, &payloadMap); err != nil {
			log.Printf("❌ Error parseando payload como map: %v", err)
			return err
		}

		// Loguear cada campo del payload con su tipo
		log.Println("📋 Campos del payload:")
		for key, value := range payloadMap {
			log.Printf("  - %s: %v (tipo: %T)", key, value, value)
		}
	}

	log.Println("========================================")
	log.Println("✅ Evento cultura procesado (solo logging)")
	log.Println("========================================")

	// TODO: Una vez conocida la estructura del payload, implementar:
	// 1. Parsear el payload a la estructura EventoCulturaPayload
	// 2. Crear un tacho fake en MySQL con id_tipo = 4 (Evento)
	// 3. Crear el bin correspondiente en MongoDB
	// 4. Asignar características por defecto
	// 5. Considerar si se necesita tabla MySQL para almacenar eventos

	fmt.Printf("Evento cultura recibido y logueado correctamente\n")
	return nil // ACK el mensaje
}
