package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publish publica un evento genérico en RabbitMQ
func Publish(ctx context.Context, routingKey string, payload interface{}) error {
	client := GetInstance()
	if client == nil {
		return fmt.Errorf("cliente de eventos no inicializado, ejecuta InitClient() primero")
	}

	if !client.IsConnected() {
		return fmt.Errorf("no hay conexión activa con RabbitMQ")
	}

	// Serializar payload a JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	// Publicar mensaje
	err = client.channel.PublishWithContext(
		ctx,
		client.exchangeName,
		routingKey,
		false, // mandatory: false = no error si no hay colas
		false, // immediate (deprecated)
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // 2 = persistente (sobrevive reinicio del broker)
			Timestamp:    time.Now(),
			AppId:        "residuos-service",
			Body:         body,
		},
	)

	if err != nil {
		return fmt.Errorf("error publicando evento: %w", err)
	}

	log.Printf("📤 Evento publicado: [%s]", routingKey)
	return nil
}

// PublishWithConfirm publica un evento y espera confirmación del broker
func PublishWithConfirm(ctx context.Context, routingKey string, payload interface{}) error {
	client := GetInstance()
	if client == nil {
		return fmt.Errorf("cliente de eventos no inicializado")
	}

	if !client.IsConnected() {
		return fmt.Errorf("no hay conexión activa con RabbitMQ")
	}

	// Habilitar confirmaciones
	if err := client.channel.Confirm(false); err != nil {
		return fmt.Errorf("error habilitando confirmaciones: %w", err)
	}

	confirms := client.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	// Serializar y publicar
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	err = client.channel.PublishWithContext(
		ctx,
		client.exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			AppId:        "residuos-service",
			Body:         body,
		},
	)

	if err != nil {
		return fmt.Errorf("error publicando evento: %w", err)
	}

	// Esperar confirmación
	select {
	case confirm := <-confirms:
		if confirm.Ack {
			log.Printf("✅ Evento confirmado: [%s]", routingKey)
			return nil
		}
		return fmt.Errorf("evento no confirmado por el broker")
	case <-ctx.Done():
		return fmt.Errorf("timeout esperando confirmación del evento")
	}
}

// PublishDirect publica un evento sin necesidad de contexto HTTP (útil para handlers de eventos)
func PublishDirect(routingKey string, payload interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return Publish(ctx, routingKey, payload)
}
