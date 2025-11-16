package eventhandlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events/schemas"
	"github.com/rabbitmq/amqp091-go"
)

// ReclamoEstadoHandler maneja notificaciones de cambios en el estado de reclamos
func ReclamoEstadoHandler(delivery amqp091.Delivery) error {
	var evento schemas.ReclamoEstadoEvent
	if err := json.Unmarshal(delivery.Body, &evento); err != nil {
		return fmt.Errorf("error al deserializar ReclamoEstadoEvent: %w", err)
	}

	log.Printf("📋 [RECLAMO ESTADO] ReclamoID=%d | TachoID=%d | Estado=%s",
		evento.ReclamoID, evento.TachoID, evento.Estado)
	log.Printf("📝 Motivo: %s", evento.Motivo)

	// Procesar según el estado del reclamo
	switch evento.Estado {
	case "Resuelto":
		return procesarReclamoResuelto(evento)
	case "Rechazado":
		return procesarReclamoRechazado(evento)
	default:
		log.Printf("⚠️ Estado de reclamo desconocido: %s", evento.Estado)
		return nil
	}
}

// procesarReclamoResuelto maneja reclamos que fueron solucionados
func procesarReclamoResuelto(evento schemas.ReclamoEstadoEvent) error {
	log.Printf("✅ [RECLAMO RESUELTO] ReclamoID=%d para TachoID=%d", evento.ReclamoID, evento.TachoID)

	// TODO: Implementación real:
	// 1. Actualizar estado del tacho en MySQL (cambiar de "en_mantenimiento" a "activo")
	// 2. Verificar tipo de reclamo original para determinar acciones específicas:
	//    - Si era "roto": validar que se haya reparado físicamente
	//    - Si era "lleno/desbordado": marcar como vaciado y resetear capacidad
	//    - Si era "mal_estado": actualizar descripción y estado operativo
	// 3. Publicar evento TachoActualizado para notificar a otros módulos
	// 4. Actualizar tabla de incidencias con fecha de resolución
	// 5. Si el tacho estaba excluido de rutas, notificar a Neo4j para incluirlo nuevamente
	// 6. Enviar notificación al usuario que reportó (integración con módulo Usuarios)

	// Mocked - Simulación
	log.Printf("   🔧 Tacho %d restaurado a estado operativo", evento.TachoID)
	log.Printf("   📤 Publicando evento 'residuos.tacho.actualizado'")
	log.Printf("   🗺️  Incluyendo tacho %d en rutas disponibles", evento.TachoID)
	log.Printf("   📧 Notificación enviada al usuario reportante")

	// Ejemplo de implementación real:
	// err := mysqlService.UpdateTachoStatus(evento.TachoID, "activo")
	// if err != nil {
	//     return fmt.Errorf("error actualizando tacho: %w", err)
	// }
	//
	// publisher.PublishTachoActualizado(evento.TachoID, "activo", evento.Motivo)
	// neo4jService.EnableTachoInRoutes(evento.TachoID)

	log.Printf("✅ Reclamo resuelto procesado: ReclamoID=%d", evento.ReclamoID)
	return nil
}

// procesarReclamoRechazado maneja reclamos que fueron descartados
func procesarReclamoRechazado(evento schemas.ReclamoEstadoEvent) error {
	log.Printf("❌ [RECLAMO RECHAZADO] ReclamoID=%d para TachoID=%d", evento.ReclamoID, evento.TachoID)

	// TODO: Implementación real:
	// 1. Verificar estado actual del tacho en MySQL
	// 2. Si el tacho estaba marcado como "en_revisión", restaurar estado anterior
	// 3. Registrar en tabla de incidencias como "Rechazado" con motivo
	// 4. Enviar notificación al usuario que reportó explicando el rechazo
	// 5. Analizar patrones de reclamos rechazados para detectar posibles reportes maliciosos
	// 6. NO publicar eventos de cambio de estado (el tacho sigue como estaba)

	// Mocked - Simulación
	log.Printf("   ⚠️ Verificando estado actual del tacho %d...", evento.TachoID)
	log.Printf("   🔄 Sin cambios en estado operativo (reclamo inválido)")
	log.Printf("   📋 Registrando rechazo: %s", evento.Motivo)
	log.Printf("   📧 Notificación enviada al usuario sobre rechazo")

	// Ejemplo de implementación real:
	// currentStatus, err := mysqlService.GetTachoStatus(evento.TachoID)
	// if err != nil {
	//     return fmt.Errorf("error consultando estado del tacho: %w", err)
	// }
	//
	// if currentStatus == "en_revision" {
	//     err = mysqlService.RevertTachoStatus(evento.TachoID)
	//     if err != nil {
	//         log.Printf("Error revirtiendo estado del tacho: %v", err)
	//     }
	// }
	//
	// incidentService.RecordRejection(evento.ReclamoID, evento.Motivo)

	log.Printf("✅ Reclamo rechazado procesado: ReclamoID=%d", evento.ReclamoID)
	return nil
}
