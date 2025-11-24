package events

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	instance         *EventClient
	once             sync.Once
	reconnectDelay   = 5 * time.Second // Reconexión cada 30 segundos
	maxInitAttempts  = 3               // Solo 3 intentos al inicio
	rabbitmqURL      string
	rabbitmqExchange string
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
		rabbitmqURL = os.Getenv("RABBITMQ_URL")
		if rabbitmqURL == "" {
			err = fmt.Errorf("RABBITMQ_URL no está configurado en las variables de entorno")
			log.Printf("❌ %v", err)
			return
		}

		rabbitmqExchange = os.Getenv("RABBITMQ_EXCHANGE")
		if rabbitmqExchange == "" {
			err = fmt.Errorf("RABBITMQ_EXCHANGE no está configurado en las variables de entorno")
			log.Printf("❌ %v", err)
			return
		}

		// Intentar conectar con reintentos limitados al inicio
		err = connectWithRetry(maxInitAttempts)
		if err != nil {
			log.Printf("❌ No se pudo conectar a RabbitMQ después de %d intentos", maxInitAttempts)
			return
		}

		// Configurar reconexión automática
		setupAutoReconnect()
	})

	return instance, err
}

// connectWithRetry intenta conectar a RabbitMQ con un número limitado de reintentos
func connectWithRetry(maxAttempts int) error {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		conn, err := amqp.Dial(rabbitmqURL)
		if err != nil {
			lastErr = err
			if attempt < maxAttempts {
				log.Printf("⚠️  Intento %d/%d fallido: %v. Reintentando en %v...", attempt, maxAttempts, err, reconnectDelay)
				time.Sleep(reconnectDelay)
				continue
			}
			return fmt.Errorf("error conectando a RabbitMQ: %w", err)
		}

		ch, err := conn.Channel()
		if err != nil {
			conn.Close()
			lastErr = err
			if attempt < maxAttempts {
				log.Printf("⚠️  Intento %d/%d fallido al crear canal: %v. Reintentando en %v...", attempt, maxAttempts, err, reconnectDelay)
				time.Sleep(reconnectDelay)
				continue
			}
			return fmt.Errorf("error creando canal: %w", err)
		}

		// Verificar que el exchange existe (modo pasivo - NO lo crea)
		if err = ch.ExchangeDeclarePassive(rabbitmqExchange, "topic", true, false, false, false, nil); err != nil {
			ch.Close()
			conn.Close()
			return fmt.Errorf("exchange '%s' no existe: %w", rabbitmqExchange, err)
		}

		instance = &EventClient{
			conn:         conn,
			channel:      ch,
			exchangeName: rabbitmqExchange,
		}

		log.Printf("✅ Cliente RabbitMQ inicializado - Exchange: %s", rabbitmqExchange)
		return nil
	}

	return lastErr
}

// setupAutoReconnect monitorea la conexión y reconecta automáticamente si se pierde
func setupAutoReconnect() {
	if instance == nil || instance.conn == nil {
		return
	}

	notifyClose := make(chan *amqp.Error)
	instance.conn.NotifyClose(notifyClose)

	go func() {
		err := <-notifyClose
		if err != nil {
			log.Printf("❌ Conexión con RabbitMQ perdida: %v", err)
			log.Println("🔄 Iniciando reconexión automática (cada 30s)...")

			// Limpiar instancia actual
			instance = nil

			// Intentar reconectar indefinidamente
			attemptReconnect()
		}
	}()
}

// attemptReconnect intenta reconectar indefinidamente hasta tener éxito
func attemptReconnect() {
	attempt := 0

	for {
		attempt++
		log.Printf("🔄 Intento de reconexión #%d...", attempt)

		conn, err := amqp.Dial(rabbitmqURL)
		if err != nil {
			log.Printf("⚠️  Reconexión fallida: %v. Reintentando en %v...", err, reconnectDelay)
			time.Sleep(reconnectDelay)
			continue
		}

		ch, err := conn.Channel()
		if err != nil {
			conn.Close()
			log.Printf("⚠️  Error creando canal: %v. Reintentando en %v...", err, reconnectDelay)
			time.Sleep(reconnectDelay)
			continue
		}

		// Verificar exchange
		if err = ch.ExchangeDeclarePassive(rabbitmqExchange, "topic", true, false, false, false, nil); err != nil {
			ch.Close()
			conn.Close()
			log.Printf("⚠️  Exchange no disponible: %v. Reintentando en %v...", err, reconnectDelay)
			time.Sleep(reconnectDelay)
			continue
		}

		instance = &EventClient{
			conn:         conn,
			channel:      ch,
			exchangeName: rabbitmqExchange,
		}

		log.Printf("✅ Reconexión exitosa después de %d intentos", attempt)

		// Reiniciar consumers
		if err := StartConsumer(); err != nil {
			log.Printf("⚠️  Error reiniciando consumers: %v", err)
		} else {
			log.Println("✅ Consumers reiniciados correctamente")
		}

		// Configurar reconexión para la nueva conexión
		setupAutoReconnect()

		return
	}
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
