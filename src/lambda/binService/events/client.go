package events

import (
	"fmt"
	"log"
	"os"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	instance *EventClient
	once     sync.Once
)

// EventClient es el cliente singleton para RabbitMQ
type EventClient struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
}

// InitClient inicializa y retorna la instancia singleton del cliente de eventos
func InitClient() (*EventClient, error) {
	var err error

	once.Do(func() {
		// Obtener configuración desde variables de entorno
		rabbitmqURL := os.Getenv("RABBITMQ_URL")
		if rabbitmqURL == "" {
			err = fmt.Errorf("RABBITMQ_URL no está configurado en las variables de entorno")
			log.Printf("❌ %v", err)
			return
		}

		exchangeName := os.Getenv("RABBITMQ_EXCHANGE")
		if exchangeName == "" {
			err = fmt.Errorf("RABBITMQ_EXCHANGE no está configurado en las variables de entorno")
			log.Printf("❌ %v", err)
			return
		}

		// Conectar a RabbitMQ
		conn, connErr := amqp.Dial(rabbitmqURL)
		if connErr != nil {
			err = connErr
			log.Printf("❌ Error conectando a RabbitMQ: %v", err)
			return
		}

		ch, chErr := conn.Channel()
		if chErr != nil {
			conn.Close()
			err = chErr
			log.Printf("❌ Error abriendo canal RabbitMQ: %v", err)
			return
		}

		// Verificar que el exchange existe (modo pasivo - NO lo crea)
		// El exchange debe ser creado por infraestructura
		if err = ch.ExchangeDeclarePassive(exchangeName, "topic", true, false, false, false, nil); err != nil {
			ch.Close()
			conn.Close()
			log.Printf("❌ Exchange '%s' no existe - debe ser creado por infraestructura: %v", exchangeName, err)
			return
		}

		instance = &EventClient{
			conn:         conn,
			channel:      ch,
			exchangeName: exchangeName,
		}

		log.Printf("✅ Cliente RabbitMQ inicializado - Exchange: %s", exchangeName)
	})

	return instance, err
}

// GetInstance retorna la instancia del cliente (debe haberse llamado InitClient primero)
func GetInstance() *EventClient {
	return instance
}

// Close cierra las conexiones de RabbitMQ
func (c *EventClient) Close() {
	if c.channel != nil {
		c.channel.Close()
		log.Println("🔌 Canal RabbitMQ cerrado")
	}
	if c.conn != nil {
		c.conn.Close()
		log.Println("🔌 Conexión RabbitMQ cerrada")
	}
}

// GetChannel retorna el canal de RabbitMQ
func (c *EventClient) GetChannel() *amqp.Channel {
	return c.channel
}

// GetExchangeName retorna el nombre del exchange configurado
func (c *EventClient) GetExchangeName() string {
	return c.exchangeName
}

// IsConnected verifica si el cliente está conectado
func (c *EventClient) IsConnected() bool {
	return c != nil && c.conn != nil && !c.conn.IsClosed()
}
