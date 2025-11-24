package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events/eventhandlers"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events/schemas"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/middleware"
	"github.com/gin-gonic/gin"
)

type RequestBody struct {
	Tipo        string  `json:"tipo" example:"incendio"`
	Descripcion string  `json:"descripcion" example:"Incendio en edificio de oficinas"`
	Latitud     float64 `json:"latitud" example:"-34.6037"`
	Longitud    float64 `json:"longitud" example:"-58.3816"`
	Prioridad   string  `json:"prioridad" example:"alta"`
}

type ResponseBody struct {
	Message     string `json:"message" example:"Emergencia enviada correctamente"`
	Tipo        string `json:"tipo" example:"incendio"`
	Descripcion string `json:"descripcion" example:"Incendio en edificio de oficinas"`
	IDAlerta    string `json:"id_alerta" example:"12345"`
}

// SendEmergencyHandler envía una emergencia
// @Summary Enviar emergencia
// @Description Registra una nueva emergencia en el sistema y publica evento a RabbitMQ
// @Tags Emergencias
// @Accept json
// @Produce json
// @Param emergencia body RequestBody true "Datos de la emergencia"
// @Success 200 {object} ResponseBody "Emergencia enviada correctamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Router /enviar-emergencia [post]
func SendEmergencyHandler(c *gin.Context) {
	var body RequestBody

	// BindJSON intenta convertir el JSON enviado a la estructura RequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	// Generar ID único para la alerta
	idAlerta := fmt.Sprintf("alerta_%d", time.Now().UnixNano())

	// Crear payload del evento usando AlertaVecinalEvent
	now := time.Now()
	payload := schemas.AlertaVecinalEvent{
		ID:             idAlerta,
		IDUsuario:      "sistema_local",
		Prioridad:      body.Prioridad,
		ScorePrioridad: getPrioridadScore(body.Prioridad),
		Estado:         "Pendiente",
		TipoEmergencia: body.Tipo,
		Origen:         "endpoint_local",
		Ubicacion: struct {
			Lat       float64 `json:"lat"`
			Lon       float64 `json:"lon"`
			Precision int     `json:"precision"`
		}{
			Lat:       body.Latitud,
			Lon:       body.Longitud,
			Precision: 100,
		},
		Adjuntos:  []interface{}{},
		Bateria:   100,
		Red:       "local",
		Timestamp: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Procesar la emergencia directamente (crea tacho fake y reclamo)
	if err := eventhandlers.ProcessAlertaVecinalDirect(payload, "EN_PROCESO", true); err != nil {
		log.Printf("❌ Error procesando emergencia: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error procesando emergencia",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Emergencia procesada: %s - %s (tacho fake creado)", idAlerta, body.Tipo)

	// Publicar evento a RabbitMQ (opcional - degradación graciosa)
	ctx := context.Background()
	if err := events.Publish(ctx, schemas.RoutingKeyAlertaPendiente, payload); err != nil {
		log.Printf("⚠️  Error publicando emergencia a RabbitMQ (continuando sin eventos): %v", err)
		// No retornar error, continuar sin publicar el evento
	} else {
		log.Printf("✅ Emergencia publicada a RabbitMQ: %s", idAlerta)
	}

	// Actualizar métricas de Prometheus
	middleware.IncrementEmergencias(body.Tipo, "zona_general")

	c.JSON(http.StatusOK, gin.H{
		"message":     "Emergencia enviada correctamente",
		"tipo":        body.Tipo,
		"descripcion": body.Descripcion,
		"id_alerta":   idAlerta,
	})
}

// getPrioridadScore convierte texto de prioridad a score numérico
func getPrioridadScore(prioridad string) int {
	scores := map[string]int{
		"baja":    1,
		"media":   2,
		"alta":    3,
		"urgente": 4,
	}
	if score, exists := scores[prioridad]; exists {
		return score
	}
	return 2 // default: media
}
