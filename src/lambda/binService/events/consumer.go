package events

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events/eventhandlers"
)

// StartConsumer inicia consumers para todas las colas configuradas en .env
func StartConsumer() error {
	client := GetInstance()
	if client == nil {
		log.Printf("⚠️  Cliente de eventos no inicializado, consumer no iniciado")
		return fmt.Errorf("cliente de eventos no inicializado")
	}

	if !client.IsConnected() {
		log.Printf("⚠️  No hay conexión a RabbitMQ, consumer no iniciado")
		return fmt.Errorf("no hay conexión a RabbitMQ")
	}

	// Obtener colas desde env (separadas por comas)
	queuesEnv := os.Getenv("RABBITMQ_QUEUES")
	if queuesEnv == "" {
		// Fallback: intentar RABBITMQ_QUEUE (singular) para retrocompatibilidad
		queuesEnv = os.Getenv("RABBITMQ_QUEUE")
		if queuesEnv == "" {
			queuesEnv = "residuos_def" // Default
		}
	}

	// Parsear colas (separadas por coma)
	queuesList := strings.Split(queuesEnv, ",")
	queues := make([]string, 0, len(queuesList))
	for _, q := range queuesList {
		q = strings.TrimSpace(q)
		if q != "" {
			queues = append(queues, q)
		}
	}

	if len(queues) == 0 {
		log.Println("⚠️  No hay colas configuradas para escuchar")
		return fmt.Errorf("no hay colas configuradas")
	}

	log.Printf("📋 Colas configuradas: %v", queues)

	// Iniciar un consumer por cada cola
	for _, queueName := range queues {
		if err := startConsumerForQueue(queueName, client); err != nil {
			log.Printf("⚠️  No se pudo iniciar consumer para %s: %v", queueName, err)
			// Continuar con las demás colas
			continue
		}
	}

	return nil
}

// startConsumerForQueue inicia un consumer para una cola específica
func startConsumerForQueue(queueName string, client *EventClient) error {
	ch := client.GetChannel()

	// Verificar que la cola existe (modo pasivo - no la crea)
	q, err := ch.QueueDeclarePassive(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Printf("❌ La cola '%s' no existe en RabbitMQ: %v", queueName, err)
		return err
	}

	log.Printf("✅ Cola '%s' encontrada (mensajes: %d, consumidores: %d)", q.Name, q.Messages, q.Consumers)

	// Vincular la cola con las routing keys que queremos escuchar
	routingKeys := eventhandlers.GetRoutingKeys()
	for _, key := range routingKeys {
		err = ch.QueueBind(
			q.Name,
			key,
			client.GetExchangeName(),
			false,
			nil,
		)
		if err != nil {
			log.Printf("❌ Error vinculando routing key '%s' a cola '%s': %v", key, q.Name, err)
			continue
		}
		log.Printf("📌 Cola '%s' vinculada con routing key: %s", q.Name, key)
	}

	// Configurar QoS (Quality of Service)
	err = ch.Qos(
		10,    // prefetch count: procesar max 10 mensajes sin ACK
		0,     // prefetch size: sin límite de bytes
		false, // global: aplicar solo a este canal
	)
	if err != nil {
		log.Printf("❌ Error configurando QoS para cola '%s': %v", queueName, err)
		return err
	}

	// Iniciar consumidor
	msgs, err := ch.Consume(
		q.Name,
		fmt.Sprintf("residuos-consumer-%s", queueName), // consumer tag único
		false, // auto-ack: false = ACK manual
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Printf("❌ Error iniciando consumer para cola '%s': %v", queueName, err)
		return err
	}

	log.Printf("🎧 Consumer iniciado, escuchando cola: %s", queueName)

	// Procesar mensajes en goroutine
	go func(queue string) {
		for d := range msgs {
			log.Printf("📥 [%s] Mensaje recibido - Routing Key: %s", queue, d.RoutingKey)
			eventhandlers.ProcessMessage(d)
		}
		log.Printf("⚠️  Consumer de cola %s finalizado", queue)
	}(queueName)

	return nil
}
