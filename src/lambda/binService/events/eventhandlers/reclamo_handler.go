package eventhandlers

import (
	"encoding/json"
	"log"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events/schemas"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ReclamoHandler procesa eventos de reclamos recibidos del módulo de Reclamos
func ReclamoHandler(d amqp.Delivery) error {
	var evento schemas.ReclamoRecibidoEvent

	// Parsear el JSON del mensaje
	if err := json.Unmarshal(d.Body, &evento); err != nil {
		log.Printf("❌ Error parseando evento de reclamo: %v", err)
		return err
	}

	log.Printf("📥 Reclamo recibido:")
	log.Printf("   Reclamo ID: %d", evento.ReclamoID)
	log.Printf("   Tacho ID: %d", evento.TachoID)
	log.Printf("   Tipo: %s", evento.TipoReclamo)
	log.Printf("   Descripción: %s", evento.Descripcion)

	// Procesar según el tipo de reclamo
	switch evento.TipoReclamo {
	case "mal_estado":
		return procesarReclamoMalEstado(evento)
	case "lleno":
		return procesarReclamoLleno(evento)
	case "roto":
		return procesarReclamoRoto(evento)
	case "desbordado":
		return procesarReclamoDesbordado(evento)
	default:
		log.Printf("⚠️  Tipo de reclamo desconocido: %s", evento.TipoReclamo)
		return nil
	}
}

// procesarReclamoMalEstado actualiza el estado del tacho a "requiere mantenimiento"
func procesarReclamoMalEstado(evento schemas.ReclamoRecibidoEvent) error {
	log.Printf("⚙️  Procesando reclamo de mal estado para tacho %d", evento.TachoID)

	// TODO: Implementar lógica de negocio
	// 1. Buscar tacho en la BD
	// 2. Actualizar estado a "requiere_mantenimiento"
	// 3. Registrar el reclamo asociado
	// 4. Posiblemente crear una tarea de mantenimiento
	// 5. Notificar a supervisores

	// Ejemplo:
	// db := config.GetDB()
	// db.Model(&models.Tacho{}).Where("id = ?", evento.TachoID).Update("estado", "requiere_mantenimiento")

	log.Printf("✅ Tacho %d marcado como 'requiere mantenimiento'", evento.TachoID)
	return nil
}

// procesarReclamoLleno actualiza la capacidad del tacho y programa recolección
func procesarReclamoLleno(evento schemas.ReclamoRecibidoEvent) error {
	log.Printf("⚙️  Procesando reclamo de tacho lleno %d", evento.TachoID)

	// TODO: Implementar lógica de negocio
	// 1. Actualizar capacidad del tacho a 100%
	// 2. Marcar como "requiere_recoleccion"
	// 3. Notificar al sistema de rutas para priorizar este tacho
	// 4. Publicar evento: residuos.tacho.lleno

	log.Printf("✅ Tacho %d marcado para recolección prioritaria", evento.TachoID)
	return nil
}

// procesarReclamoRoto marca el tacho como fuera de servicio
func procesarReclamoRoto(evento schemas.ReclamoRecibidoEvent) error {
	log.Printf("⚙️  Procesando reclamo de tacho roto %d", evento.TachoID)

	// TODO: Implementar lógica de negocio
	// 1. Marcar tacho como "fuera_de_servicio"
	// 2. Excluir del sistema de rutas
	// 3. Crear ticket de mantenimiento urgente
	// 4. Notificar a equipo de mantenimiento

	log.Printf("✅ Tacho %d marcado como fuera de servicio", evento.TachoID)
	return nil
}

// procesarReclamoDesbordado similar a lleno pero con mayor prioridad
func procesarReclamoDesbordado(evento schemas.ReclamoRecibidoEvent) error {
	log.Printf("⚙️  Procesando reclamo de tacho desbordado %d (URGENTE)", evento.TachoID)

	// TODO: Implementar lógica de negocio
	// 1. Marcar como emergencia
	// 2. Prioridad MÁXIMA en sistema de rutas
	// 3. Notificación inmediata a supervisores
	// 4. Publicar evento de emergencia

	log.Printf("✅ Tacho %d marcado como emergencia - desbordado", evento.TachoID)
	return nil
}

// RecoleccionHandler procesa eventos de recolección completada del módulo de Conductores
func RecoleccionHandler(d amqp.Delivery) error {
	var evento schemas.RecoleccionCompletadaEvent

	if err := json.Unmarshal(d.Body, &evento); err != nil {
		log.Printf("❌ Error parseando evento de recolección: %v", err)
		return err
	}

	log.Printf("📥 Recolección completada:")
	log.Printf("   Tacho ID: %d", evento.TachoID)
	log.Printf("   Conductor ID: %d", evento.ConductorID)
	log.Printf("   Fecha vaciado: %s", evento.FechaVaciado.Format("2006-01-02 15:04:05"))

	// TODO: Implementar lógica de negocio
	// 1. Actualizar capacidad del tacho a 0%
	// 2. Actualizar última fecha de vaciado
	// 3. Cambiar estado a "operativo"
	// 4. Registrar en historial de recolecciones

	log.Printf("✅ Tacho %d actualizado - capacidad reseteada", evento.TachoID)
	return nil
}
