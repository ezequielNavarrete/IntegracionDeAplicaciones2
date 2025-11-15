package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events/schemas"
)

// Ejemplo de cómo publicar eventos desde tus servicios

// generateEventID genera un ID único para el evento (UUID v4 simple)
func generateEventID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// PublishTachoCreado publica un evento cuando se crea un nuevo tacho
func PublishTachoCreado(tachoID int, capacidad float64, ubicacion string, zonaID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	evento := schemas.TachoCreadoEvent{
		BaseEvent: schemas.BaseEvent{
			EventID:   generateEventID(),
			Timestamp: time.Now(),
			ModuloID:  "residuos-service",
		},
		TachoID:   tachoID,
		Capacidad: capacidad,
		Ubicacion: ubicacion,
		ZonaID:    zonaID,
	}

	if err := events.Publish(ctx, schemas.RoutingKeyTachoCreado, evento); err != nil {
		log.Printf("❌ Error publicando evento TachoCreado: %v", err)
		return err
	}

	log.Printf("✅ Evento TachoCreado publicado para tacho %d", tachoID)
	return nil
}

// PublishTachoEliminado publica un evento cuando se elimina un tacho
func PublishTachoEliminado(tachoID int, motivo string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	evento := schemas.TachoEliminadoEvent{
		BaseEvent: schemas.BaseEvent{
			EventID:   generateEventID(),
			Timestamp: time.Now(),
			ModuloID:  "residuos-service",
		},
		TachoID: tachoID,
		Motivo:  motivo,
	}

	if err := events.Publish(ctx, schemas.RoutingKeyTachoEliminado, evento); err != nil {
		log.Printf("❌ Error publicando evento TachoEliminado: %v", err)
		return err
	}

	log.Printf("✅ Evento TachoEliminado publicado para tacho %d", tachoID)
	return nil
}

// PublishTachoLleno publica un evento cuando un tacho está lleno
func PublishTachoLleno(tachoID int, capacidadActual float64, zonaID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	evento := schemas.TachoLlenoEvent{
		BaseEvent: schemas.BaseEvent{
			EventID:   generateEventID(),
			Timestamp: time.Now(),
			ModuloID:  "residuos-service",
		},
		TachoID:         tachoID,
		CapacidadActual: capacidadActual,
		UltimaFecha:     time.Now(),
		ZonaID:          zonaID,
	}

	if err := events.Publish(ctx, schemas.RoutingKeyTachoLleno, evento); err != nil {
		log.Printf("❌ Error publicando evento TachoLleno: %v", err)
		return err
	}

	log.Printf("✅ Evento TachoLleno publicado para tacho %d (%.2f%%)", tachoID, capacidadActual)
	return nil
}

// PublishTachoActualizado publica un evento cuando se actualiza un tacho
func PublishTachoActualizado(tachoID int, capacidad float64, nuevaZonaID int, detalles string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	evento := schemas.TachoActualizadoEvent{
		BaseEvent: schemas.BaseEvent{
			EventID:   generateEventID(),
			Timestamp: time.Now(),
			ModuloID:  "residuos-service",
		},
		TachoID:     tachoID,
		Capacidad:   capacidad,
		NuevaZonaID: nuevaZonaID,
		Detalles:    detalles,
	}

	if err := events.Publish(ctx, schemas.RoutingKeyTachoActualizado, evento); err != nil {
		log.Printf("❌ Error publicando evento TachoActualizado: %v", err)
		return err
	}

	log.Printf("✅ Evento TachoActualizado publicado para tacho %d", tachoID)
	return nil
}

// PublishAlertaResuelta publica un evento cuando se resuelve una alerta vecinal
func PublishAlertaResuelta(emergenciaID string, rutasPerjudicadas int, lat, lng float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	evento := schemas.AlertaResueltaEvent{
		EventID:           generateEventID(),
		EmergenciaID:      emergenciaID,
		Estado:            "Resuelto",
		RutasPerjudicadas: rutasPerjudicadas,
		Lng:               lng,
		Lat:               lat,
		Timestamp:         time.Now(),
		ModuloID:          "residuos-service",
	}

	if err := events.Publish(ctx, schemas.RoutingKeyAlertaResuelta, evento); err != nil {
		log.Printf("❌ Error publicando evento AlertaResuelta: %v", err)
		return err
	}

	log.Printf("✅ Evento AlertaResuelta publicado para emergencia %s (rutas afectadas: %d)", emergenciaID, rutasPerjudicadas)
	return nil
}

// Ejemplo de uso en tus handlers existentes:
/*

// En handlers/tachos_handler.go

func CrearTacho(c *gin.Context) {
	var tacho models.Tacho
	if err := c.ShouldBindJSON(&tacho); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Guardar en BD
	if err := db.Create(&tacho).Error; err != nil {
		c.JSON(500, gin.H{"error": "Error creando tacho"})
		return
	}

	// 🔥 Publicar evento
	go services.PublishTachoCreado(
		tacho.ID,
		tacho.Capacidad,
		tacho.Ubicacion,
		tacho.ZonaID,
	)

	c.JSON(201, tacho)
}

func EliminarTacho(c *gin.Context) {
	id := c.Param("id")

	// Eliminar de BD
	if err := db.Delete(&models.Tacho{}, id).Error; err != nil {
		c.JSON(500, gin.H{"error": "Error eliminando tacho"})
		return
	}

	// 🔥 Publicar evento
	go services.PublishTachoEliminado(
		parseID(id),
		"Eliminado por usuario",
	)

	c.JSON(200, gin.H{"message": "Tacho eliminado"})
}

*/

// parseID convierte string a int (helper function)
func parseID(id string) int {
	var result int
	fmt.Sscanf(id, "%d", &result)
	return result
}
