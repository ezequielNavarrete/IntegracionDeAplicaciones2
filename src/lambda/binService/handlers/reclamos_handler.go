package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events/schemas"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetAllReclamosHandler obtiene todos los reclamos
// @Summary Obtener todos los reclamos
// @Description Devuelve una lista de todos los reclamos. Opcionalmente se puede filtrar por estado usando el query parameter ?estado=PENDIENTE
// @Tags Reclamos
// @Accept json
// @Produce json
// @Param estado query string false "Filtrar por estado (ej: PENDIENTE, ESPERA_INFO, RECHAZADO, RESUELTO)"
// @Success 200 {array} services.Reclamo "Lista de reclamos"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /reclamos [get]
func GetAllReclamosHandler(c *gin.Context) {
	// Obtener filtro opcional de estado desde query params
	estadoFiltro := c.Query("estado")

	reclamos, err := services.GetAllReclamos(estadoFiltro)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"reclamos": reclamos,
		"total":    len(reclamos),
	}

	// Agregar el filtro aplicado en la respuesta si existe
	if estadoFiltro != "" {
		response["filtro_estado"] = estadoFiltro
	}

	c.JSON(http.StatusOK, response)
}

// GetReclamoByIDHandler obtiene un reclamo específico por ID
// @Summary Obtener un reclamo por ID
// @Description Devuelve un reclamo específico
// @Tags Reclamos
// @Accept json
// @Produce json
// @Param id path int true "ID del reclamo"
// @Success 200 {object} services.Reclamo "Reclamo encontrado"
// @Failure 400 {object} map[string]string "ID inválido"
// @Failure 404 {object} map[string]string "Reclamo no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /reclamos/{id} [get]
func GetReclamoByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	reclamoID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	reclamo, err := services.GetReclamoByID(reclamoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reclamo no encontrado"})
		return
	}

	c.JSON(http.StatusOK, reclamo)
}

// CreateReclamoHandler crea un nuevo reclamo
// @Summary Crear un nuevo reclamo
// @Description Crea un reclamo con la información proporcionada
// @Tags Reclamos
// @Accept json
// @Produce json
// @Param reclamo body services.CreateReclamoRequest true "Datos del reclamo"
// @Success 201 {object} services.CreateReclamoResponse "Reclamo creado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /reclamos [post]
func CreateReclamoHandler(c *gin.Context) {
	var request services.CreateReclamoRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	response, err := services.CreateReclamo(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// DeleteReclamoHandler elimina un reclamo
// @Summary Eliminar un reclamo
// @Description Elimina un reclamo por su ID
// @Tags Reclamos
// @Accept json
// @Produce json
// @Param id path int true "ID del reclamo"
// @Success 200 {object} map[string]string "Reclamo eliminado exitosamente"
// @Failure 400 {object} map[string]string "ID inválido"
// @Failure 404 {object} map[string]string "Reclamo no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /reclamos/{id} [delete]
func DeleteReclamoHandler(c *gin.Context) {
	idStr := c.Param("id")
	reclamoID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := services.DeleteReclamo(reclamoID); err != nil {
		if err.Error() == "reclamo not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reclamo no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reclamo eliminado exitosamente"})
}

// UpdateReclamoEstadoHandler actualiza el estado de un reclamo y publica evento a RabbitMQ
// @Summary Actualizar estado de un reclamo
// @Description Actualiza el estado de un reclamo específico. Estados permitidos: ESPERA_INFO, RECHAZADO, RESUELTO
// @Tags Reclamos
// @Accept json
// @Produce json
// @Param id path int true "ID del reclamo"
// @Param request body object{estado=string,comentario=string} true "Estado y comentario opcional"
// @Success 200 {object} map[string]string "Estado actualizado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 404 {object} map[string]string "Reclamo no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /reclamos/{id}/estado [put]
func UpdateReclamoEstadoHandler(c *gin.Context) {
	idStr := c.Param("id")
	reclamoID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var body struct {
		Estado     string `json:"estado" binding:"required"`
		Comentario string `json:"comentario"` // Opcional
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Estado requerido"})
		return
	}

	// Validar que el estado sea uno de los permitidos
	estadosPermitidos := map[string]bool{
		"ESPERA_INFO": true,
		"RECHAZADO":   true,
		"RESUELTO":    true,
	}

	if !estadosPermitidos[body.Estado] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Estado inválido. Estados permitidos: ESPERA_INFO, RECHAZADO, RESUELTO",
		})
		return
	}

	// Primero obtener el reclamo para verificar que existe y obtener su ID externo
	reclamo, err := services.GetReclamoByID(reclamoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reclamo no encontrado"})
		return
	}

	// Actualizar estado en MySQL
	if err := services.UpdateReclamoEstado(reclamoID, body.Estado); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar estado del reclamo"})
		return
	}

	// Publicar evento a RabbitMQ usando el reclamo que ya obtuvimos
	if err := publishReclamoEstadoEvent(c, reclamo, body.Estado, body.Comentario); err != nil {
		log.Printf("⚠️  [UpdateReclamoEstado] Error publicando evento a RabbitMQ: %v (el estado se actualizó en MySQL)", err)
		// No retornamos error porque el estado ya se actualizó en MySQL
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Estado actualizado exitosamente",
		"estado":  body.Estado,
	})
}

// publishReclamoEstadoEvent publica el evento de cambio de estado a RabbitMQ
func publishReclamoEstadoEvent(c *gin.Context, reclamo *services.Reclamo, estado, comentario string) error {
	// Usar el id_reclamo_externo si existe, sino el id interno
	idParaPublicar := reclamo.IDReclamo
	if reclamo.IDReclamoExterno != nil && *reclamo.IDReclamoExterno > 0 {
		idParaPublicar = *reclamo.IDReclamoExterno
		log.Printf("📤 [PublishReclamoEstado] Usando ID externo para publicar: %d (ID interno: %d)", idParaPublicar, reclamo.IDReclamo)
	} else {
		log.Printf("📤 [PublishReclamoEstado] Usando ID interno para publicar: %d (no tiene ID externo)", reclamo.IDReclamo)
	}

	// Determinar routing key según el estado
	var routingKey string
	switch estado {
	case "RESUELTO":
		routingKey = schemas.RoutingKeyReclamoResueltoPub
	case "RECHAZADO":
		routingKey = schemas.RoutingKeyReclamoRechazadoPub
	case "ESPERA_INFO":
		routingKey = schemas.RoutingKeyReclamoEsperaInfoPub
	default:
		return nil // No publicar si no es un estado conocido
	}

	// Crear payload del evento con el ID externo
	payload := schemas.ReclamoEstadoCambiadoPayload{
		IDReclamo:      idParaPublicar, // Usar ID externo del sistema de Reclamos
		Comentario:     comentario,
		Estado:         estado,
		FechaRespuesta: time.Now(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Crear envelope estándar
	envelope := schemas.EventEnvelope{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Source:    "residuos",
		Topic:     routingKey,
		Payload:   payloadBytes,
	}

	// Publicar evento usando el contexto de la petición HTTP
	ctx := c.Request.Context()
	if err := events.Publish(ctx, routingKey, envelope); err != nil {
		return err
	}

	log.Printf("✅ [UpdateReclamoEstado] Evento publicado - Routing Key: %s, Reclamo ID: %d, Estado: %s",
		routingKey, reclamo.IDReclamo, estado)

	return nil
}
