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

	// 1. Buscar y actualizar tacho fake en MongoDB por id_evento (ponerlo fuera de servicio)
	tachosCollection, err := config.GetMongoCollection("tachos")
	if err != nil {
		log.Printf("❌ [EventoCulturaCancelado] Error obteniendo colección: %v", err)
		return fmt.Errorf("error obteniendo colección: %v", err)
	}
	
	filter := bson.M{"id_evento": payload.IDEvento}
	update := bson.M{
		"$set": bson.M{
			"estado": "fuera_de_servicio",
			"motivo_cancelacion": payload.Status,
		},
	}
	
	ctx := context.Background()
	result, err := tachosCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Printf("❌ [EventoCulturaCancelado] Error actualizando tacho fake en MongoDB: %v", err)
		return fmt.Errorf("error actualizando tacho fake en MongoDB: %v", err)
	}

	tachoIDsMongo := []int{} // Para trackear qué tachos actualizar en MySQL

	if result.ModifiedCount > 0 {
		log.Printf("✅ [EventoCulturaCancelado] Tacho(s) fake puesto(s) fuera de servicio en MongoDB: %d (id_evento: %s)", 
			result.ModifiedCount, payload.IDEvento)
			
		// Obtener los id_mongo de los tachos actualizados para actualizar MySQL
		cursor, err := tachosCollection.Find(ctx, filter)
		if err == nil {
			defer cursor.Close(ctx)
			for cursor.Next(ctx) {
				var tacho struct {
					IDMongo int `bson:"_id"`
				}
				if err := cursor.Decode(&tacho); err == nil {
					tachoIDsMongo = append(tachoIDsMongo, tacho.IDMongo)
				}
			}
		}
	} else {
		log.Printf("⚠️  [EventoCulturaCancelado] No se encontró tacho fake con id_evento: %s", payload.IDEvento)
		
		// Intentar buscar por coordenadas como fallback
		coordFilter := bson.M{
			"lat": payload.Longitude, // MongoDB tiene invertidas las coordenadas
			"lon": payload.Latitude,  // MongoDB tiene invertidas las coordenadas
			"id_tipo": 4, // Tachos fake de eventos culturales
		}
		
		coordResult, coordErr := tachosCollection.UpdateMany(ctx, coordFilter, update)
		if coordErr != nil {
			log.Printf("❌ [EventoCulturaCancelado] Error buscando por coordenadas: %v", coordErr)
			return fmt.Errorf("error buscando por coordenadas: %v", coordErr)
		}
		
		if coordResult.ModifiedCount > 0 {
			log.Printf("✅ [EventoCulturaCancelado] Tacho(s) fake puesto(s) fuera de servicio por coordenadas en MongoDB: %d", coordResult.ModifiedCount)
			
			// Obtener los id_mongo de los tachos actualizados
			cursor, err := tachosCollection.Find(ctx, coordFilter)
			if err == nil {
				defer cursor.Close(ctx)
				for cursor.Next(ctx) {
					var tacho struct {
						IDMongo int `bson:"_id"`
					}
					if err := cursor.Decode(&tacho); err == nil {
						tachoIDsMongo = append(tachoIDsMongo, tacho.IDMongo)
					}
				}
			}
		} else {
			log.Printf("⚠️  [EventoCulturaCancelado] No se encontró tacho fake ni por id_evento ni por coordenadas")
		}
	}

	// 2. Actualizar en MySQL los tachos correspondientes (id_estado = 4: fuera de servicio)
	if len(tachoIDsMongo) > 0 {
		log.Printf("📝 [EventoCulturaCancelado] Actualizando %d tacho(s) en MySQL...", len(tachoIDsMongo))
		
		// Construir query para actualizar múltiples tachos
		query := "UPDATE Tacho SET id_estado = 4 WHERE id_mongo IN ("
		for i := range tachoIDsMongo {
			if i > 0 {
				query += ", "
			}
			query += "?"
		}
		query += ")"
		
		// Convertir slice a []interface{} para los parámetros
		params := make([]interface{}, len(tachoIDsMongo))
		for i, id := range tachoIDsMongo {
			params[i] = id
		}
		
		if result := config.DB.Exec(query, params...); result.Error() != nil {
			log.Printf("⚠️  [EventoCulturaCancelado] Error actualizando MySQL: %v", result.Error())
			// No retornamos error porque MongoDB ya se actualizó
		} else {
			log.Printf("✅ [EventoCulturaCancelado] Tachos actualizados en MySQL exitosamente")
		}
	}

	log.Println("========================================")
	log.Println("✅ Evento cultura cancelado procesado")
	log.Println("========================================")

	return nil
}
