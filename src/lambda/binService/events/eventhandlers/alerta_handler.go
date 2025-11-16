package eventhandlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events/schemas"
	"github.com/rabbitmq/amqp091-go"
)

// AlertaVecinalHandler maneja eventos de emergencias sobre tachos dañados
func AlertaVecinalHandler(delivery amqp091.Delivery) error {
	var alerta schemas.AlertaVecinalEvent
	if err := json.Unmarshal(delivery.Body, &alerta); err != nil {
		return fmt.Errorf("error al deserializar AlertaVecinalEvent: %w", err)
	}

	log.Printf("🚨 [ALERTA VECINAL] ID=%s | Estado=%s | Tipo=%s | Prioridad=%s (Score: %d)",
		alerta.ID, alerta.Estado, alerta.TipoEmergencia, alerta.Prioridad, alerta.ScorePrioridad)
	log.Printf("📍 Ubicación: Lat=%.6f, Lon=%.6f (Precisión: %dm)",
		alerta.Ubicacion.Lat, alerta.Ubicacion.Lon, alerta.Ubicacion.Precision)

	// Procesar según el estado de la alerta
	switch alerta.Estado {
	case "Pendiente":
		return procesarAlertaPendiente(alerta)
	case "Resuelta":
		return procesarAlertaResuelta(alerta)
	case "Cancelada":
		return procesarAlertaCancelada(alerta)
	default:
		log.Printf("⚠️ Estado de alerta desconocido: %s", alerta.Estado)
		return nil
	}
}

// procesarAlertaPendiente maneja alertas nuevas que requieren atención
func procesarAlertaPendiente(alerta schemas.AlertaVecinalEvent) error {
	log.Printf("🔴 [ALERTA PENDIENTE] Procesando emergencia ID=%s", alerta.ID)

	// TODO: Implementación real:
	// 1. Buscar tachos cercanos a la ubicación (lat/lon) usando base de datos
	// 2. Marcar tachos dentro del radio de impacto como "fuera de servicio"
	// 3. Notificar al servicio de rutas para recalcular trayectos
	// 4. Si ScorePrioridad > 80, enviar notificación urgente a operadores
	// 5. Registrar la emergencia en la tabla de incidencias

	// Mocked - Simulación
	log.Printf("   🚫 Tacho hipotético cercano marcado fuera de servicio")
	log.Printf("   🗺️  Rutas afectadas: Recalculando...")
	if alerta.ScorePrioridad > 80 {
		log.Printf("   🚨 URGENTE: Notificación enviada a operadores (Score alto)")
	}

	log.Printf("✅ Alerta pendiente procesada: %s", alerta.ID)
	return nil
}

// procesarAlertaResuelta maneja alertas que fueron resueltas por emergencias
func procesarAlertaResuelta(alerta schemas.AlertaVecinalEvent) error {
	log.Printf("🟢 [ALERTA RESUELTA] Emergencia ID=%s finalizada", alerta.ID)

	// TODO: Implementación real:
	// 1. Buscar tachos que fueron marcados como fuera de servicio por esta emergencia
	// 2. Restaurar estado "activo" de los tachos afectados
	// 3. Publicar evento AlertaResueltaEvent para notificar a Movilidad
	// 4. Actualizar tabla de incidencias con fecha de resolución
	// 5. Recalcular rutas optimizadas con los tachos restaurados

	// Mocked - Simulación
	log.Printf("   ✅ Tachos restaurados a estado operativo")
	log.Printf("   📤 Publicando evento 'residuos.alertavecinal.resuelta' para Movilidad")
	log.Printf("   🗺️  Rutas recalculadas con tachos restaurados")

	// TODO: Llamar a services.PublishAlertaResuelta(alerta.ID, cantidadRutas)

	log.Printf("✅ Alerta resuelta procesada: %s", alerta.ID)
	return nil
}

// procesarAlertaCancelada maneja alertas canceladas (falsa alarma)
func procesarAlertaCancelada(alerta schemas.AlertaVecinalEvent) error {
	log.Printf("⚪ [ALERTA CANCELADA] Emergencia ID=%s descartada", alerta.ID)

	// TODO: Implementación real:
	// 1. Verificar si se habían marcado tachos como fuera de servicio
	// 2. Restaurar tachos a estado normal si corresponde
	// 3. Actualizar tabla de incidencias como "Cancelada"
	// 4. No publicar eventos a otros módulos (fue falsa alarma)

	// Mocked - Simulación
	log.Printf("   ⚠️ Revisando si hubo impacto en tachos...")
	log.Printf("   ✅ Sin cambios requeridos (falsa alarma)")

	log.Printf("✅ Alerta cancelada procesada: %s", alerta.ID)
	return nil
}
