package eventhandlers

import (
	"log"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events/schemas"
	amqp "github.com/rabbitmq/amqp091-go"
)

// HandlerFunc es la firma de una función manejadora de eventos
type HandlerFunc func(amqp.Delivery) error

// GetHandlers retorna el mapeo de routing keys a sus handlers correspondientes
func GetHandlers() map[string]HandlerFunc {
	return map[string]HandlerFunc{
		// Eventos de Reclamos (antiguos - tachos con problemas)
		schemas.RoutingKeyReclamoMalEstado:  ReclamoHandler,
		schemas.RoutingKeyReclamoLleno:      ReclamoHandler,
		schemas.RoutingKeyReclamoRoto:       ReclamoHandler,
		schemas.RoutingKeyReclamoDesbordado: ReclamoHandler,

		// Eventos de Reclamos (nuevos - formato envelope estándar)
		schemas.RoutingKeyReclamoResiduoCreado:   ReclamoResiduoHandler,
		schemas.RoutingKeyReclamoResiduoDerivado: ReclamoResiduoHandler, // Reutiliza el mismo handler

		// Eventos de Conductores (recolecciones completadas)
		schemas.RoutingKeyRecoleccionCompletada: RecoleccionHandler,

		// Eventos de Cultura (solo logging por ahora)
		schemas.RoutingKeyEventoCulturaCrear:    EventoCulturaHandler,
		schemas.RoutingKeyEventoCulturaCancelar: EventoCulturaCanceladoHandler,

		// Nuevos handlers de otros módulos
		schemas.RoutingKeyAlertaPendiente: AlertaVecinalHandler, // De Emergencia
		// De Reclamos (rechazo)
	}
}

// GetRoutingKeys retorna todas las routing keys que el módulo debe escuchar
func GetRoutingKeys() []string {
	return []string{
		// Reclamos específicos (antiguos)
		schemas.RoutingKeyReclamoMalEstado,
		schemas.RoutingKeyReclamoLleno,
		schemas.RoutingKeyReclamoRoto,
		schemas.RoutingKeyReclamoDesbordado,

		// Reclamos de residuos (nuevos - formato envelope)
		schemas.RoutingKeyReclamoResiduoCreado,
		schemas.RoutingKeyReclamoResiduoDerivado,

		// Patrón wildcard para cualquier reclamo de tachos (backup)
		"reclamos.tacho.#",

		// Recolecciones
		schemas.RoutingKeyRecoleccionCompletada,

		// Cultura
		schemas.RoutingKeyEventoCulturaCrear,
		schemas.RoutingKeyEventoCulturaCancelar,

		// Emergencias
		schemas.RoutingKeyAlertaPendiente,
	}
}

// ProcessMessage enruta el mensaje al handler correcto según la routing key
func ProcessMessage(d amqp.Delivery) {
	log.Printf("📨 Mensaje recibido: [%s]", d.RoutingKey)

	handlers := GetHandlers()

	// Buscar handler específico para esta routing key
	handler, exists := handlers[d.RoutingKey]
	if !exists {
		log.Printf("⚠️  No hay handler específico para routing key: %s", d.RoutingKey)
		log.Printf("   Contenido: %s", string(d.Body))
		// ACK del mensaje aunque no tengamos handler (evitar requeue infinito)
		d.Ack(false)
		return
	}

	// Ejecutar el handler
	if err := handler(d); err != nil {
		log.Printf("❌ Error procesando mensaje [%s]: %v", d.RoutingKey, err)
		log.Printf("🗑️  Mensaje descartado (NO se reencola): %s", string(d.Body))
		
		// Hacer NACK sin reencolar (requeue = false)
		// Esto descarta el mensaje definitivamente
		d.Nack(false, false) // multiple=false, requeue=false
		return
	}

	// ACK: confirmar que el mensaje fue procesado exitosamente
	if err := d.Ack(false); err != nil {
		log.Printf("❌ Error haciendo ACK del mensaje: %v", err)
	} else {
		log.Printf("✅ Mensaje procesado correctamente: [%s]", d.RoutingKey)
	}
}
