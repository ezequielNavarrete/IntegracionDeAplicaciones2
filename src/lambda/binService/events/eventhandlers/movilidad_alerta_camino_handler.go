package eventhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

// MovilidadAlertaCaminoEnvelope representa el mensaje recibido desde Movilidad
// Ejemplo de mensaje:
// {
//   "topic": "movilidad.alerta.camino",
//   "payload": {
//     "tripId": "emergency_...",
//     "origin": {"coordinates": {"lng": -58.38, "lat": -34.60}},
//     "destination": {"coordinates": {"lng": -58.38, "lat": -34.60}}
//   }
// }
// Nota: No usamos el envelope estándar del repo; este handler parsea el esquema indicado arriba.

type movilidadAlertaCaminoEnvelope struct {
	Topic   string                       `json:"topic"`
	Payload movilidadAlertaCaminoPayload `json:"payload"`
}

type movilidadAlertaCaminoPayload struct {
	TripID      string            `json:"tripId"`
	Origin      movilidadEndpoint `json:"origin"`
	Destination movilidadEndpoint `json:"destination"`
}

type movilidadEndpoint struct {
	Coordinates coords `json:"coordinates"`
}

type coords struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

type storedRoadblock struct {
	TripID      string    `json:"tripId"`
	Origin      coords    `json:"origin"`
	Destination coords    `json:"destination"`
	CreatedAt   time.Time `json:"createdAt"`
}

// MovilidadAlertaCaminoHandler procesa la alerta de camino bloqueado y guarda coordenadas en Redis
func MovilidadAlertaCaminoHandler(d amqp.Delivery) error {
	log.Println("========================================")
	log.Printf("🚧 Evento recibido: %s", d.RoutingKey)
	log.Println("========================================")

	var env movilidadAlertaCaminoEnvelope
	if err := json.Unmarshal(d.Body, &env); err != nil {
		log.Printf("❌ [MovilidadAlertaCamino] Error parseando mensaje: %v", err)
		return fmt.Errorf("error parsing message: %w", err)
	}

	// Validaciones mínimas
	if env.Payload.TripID == "" {
		return fmt.Errorf("payload.tripId es requerido")
	}

	// Preparar valor a guardar
	record := storedRoadblock{
		TripID:      env.Payload.TripID,
		Origin:      env.Payload.Origin.Coordinates,
		Destination: env.Payload.Destination.Coordinates,
		CreatedAt:   time.Now().UTC(),
	}

	b, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("error marshaling record: %w", err)
	}

	if config.RedisClient == nil {
		return fmt.Errorf("redis client not available")
	}

	// TTL configurable por env, default 1 hora
	ttl := time.Hour
	if v := os.Getenv("MOVILIDAD_ROADBLOCK_TTL_SECONDS"); v != "" {
		if dur, perr := time.ParseDuration(v + "s"); perr == nil {
			ttl = dur
		}
	}

	key := fmt.Sprintf("movilidad:camino:bloqueado:%s", env.Payload.TripID)
	if err := config.RedisClient.Set(context.Background(), key, b, ttl).Err(); err != nil {
		log.Printf("❌ [MovilidadAlertaCamino] Error guardando en Redis: %v", err)
		return fmt.Errorf("redis set failed: %w", err)
	}

	log.Printf("✅ Camino bloqueado guardado en Redis | key=%s | ttl=%s", key, ttl.String())
	log.Printf("📍 Origen: lat=%.6f, lng=%.6f | Destino: lat=%.6f, lng=%.6f",
		record.Origin.Lat, record.Origin.Lng, record.Destination.Lat, record.Destination.Lng)

	return nil
}
