package eventhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events/schemas"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson"
)

// EventoCulturaCanceladoHandler maneja eventos de cancelación de eventos culturales
// Elimina el tacho fake asociado al evento usando id_evento o coordenadas
func EventoCulturaCanceladoHandler(d amqp.Delivery) error {
	log.Println("========================================")
	log.Println("🚫 Evento cultura.evento.cancelado recibido")
	log.Println("========================================")

	log.Printf("📄 Body raw: %s", string(d.Body))

	// Parsear envelope estándar
	var envelope schemas.EventEnvelope
	if err := json.Unmarshal(d.Body, &envelope); err != nil {
		log.Printf("❌ [EventoCulturaCancelado] Error parseando envelope: %v", err)
		return fmt.Errorf("error parsing envelope: %v", err)
	}

	log.Printf("📦 Envelope ID: %s", envelope.ID)
	log.Printf("📦 Source: %s", envelope.Source)
	log.Printf("📦 Topic: %s", envelope.Topic)

	// Parsear payload
	var payload schemas.EventoCulturaCanceladoPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		log.Printf("❌ [EventoCulturaCancelado] Error parseando payload: %v", err)
		return fmt.Errorf("error parsing payload: %v", err)
	}

	log.Printf("📋 ID Evento: %s", payload.IDEvento)
	log.Printf("📋 Coordenadas: lat=%v, lng=%v", payload.Latitude, payload.Longitude)
	log.Printf("📋 Status: %s", payload.Status)

	if payload.IDEvento == "" {
		log.Println("❌ [EventoCulturaCancelado] ID de evento faltante")
		return fmt.Errorf("id_evento es requerido")
	}

	// Buscar y eliminar tacho fake en MongoDB por id_evento
	tachosCollection, err := config.GetMongoCollection("tachos")
	if err != nil {
		log.Printf("❌ [EventoCulturaCancelado] Error obteniendo colección: %v", err)
		return fmt.Errorf("error obteniendo colección: %v", err)
	}
	
	filter := bson.M{"id_evento": payload.IDEvento}
	
	ctx := context.Background()
	result, err := tachosCollection.DeleteMany(ctx, filter)
	if err != nil {
		log.Printf("❌ [EventoCulturaCancelado] Error eliminando tacho fake: %v", err)
		return fmt.Errorf("error eliminando tacho fake: %v", err)
	}

	if result.DeletedCount > 0 {
		log.Printf("✅ [EventoCulturaCancelado] Tacho(s) fake eliminado(s): %d (id_evento: %s)", 
			result.DeletedCount, payload.IDEvento)
	} else {
		log.Printf("⚠️  [EventoCulturaCancelado] No se encontró tacho fake con id_evento: %s", payload.IDEvento)
		
		// Intentar buscar por coordenadas como fallback
		coordFilter := bson.M{
			"lat": payload.Longitude, // MongoDB tiene invertidas las coordenadas
			"lon": payload.Latitude,  // MongoDB tiene invertidas las coordenadas
			"id_tipo": 4, // Tachos fake de eventos culturales
		}
		
		coordResult, coordErr := tachosCollection.DeleteMany(ctx, coordFilter)
		if coordErr != nil {
			log.Printf("❌ [EventoCulturaCancelado] Error buscando por coordenadas: %v", coordErr)
			return fmt.Errorf("error buscando por coordenadas: %v", coordErr)
		}
		
		if coordResult.DeletedCount > 0 {
			log.Printf("✅ [EventoCulturaCancelado] Tacho(s) fake eliminado(s) por coordenadas: %d", coordResult.DeletedCount)
		} else {
			log.Printf("⚠️  [EventoCulturaCancelado] No se encontró tacho fake ni por id_evento ni por coordenadas")
		}
	}

	log.Println("========================================")
	log.Println("✅ Evento cultura cancelado procesado")
	log.Println("========================================")

	return nil
}
