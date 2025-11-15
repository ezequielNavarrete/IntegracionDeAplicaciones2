package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/middleware"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
)

// CreateTachoHandler crea un nuevo tacho en MySQL y Neo4j
// @Summary Crear un nuevo tacho
// @Description Crea un tacho guardándolo tanto en MySQL como en Neo4j
// @Tags Tachos
// @Accept json
// @Produce json
// @Param tacho body services.CreateTachoRequest true "Datos del tacho a crear"
// @Success 201 {object} services.CreateTachoResponse "Tacho creado exitosamente"
// @Failure 400 {object} map[string]string "Datos de entrada inválidos"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /tachos [post]
func CreateTachoHandler(c *gin.Context) {
	var request services.CreateTachoRequest

	// Validar y bind del JSON
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	// Crear el tacho usando el servicio
	response, err := services.CreateTacho(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Actualizar métricas después de crear exitosamente
	// Incrementar contador de tachos para la zona (usando neighborhood)
	middleware.UpdateTachoCapacidad(string(rune(response.TachoID)), string(rune(request.Neighborhood)), request.Capacidad)

	// 🔥 Publicar evento de creación a RabbitMQ
	go services.PublishTachoCreado(
		response.TachoID,
		request.Capacidad,
		fmt.Sprintf("Barrio %d (Lat: %.6f, Lon: %.6f)", request.Neighborhood, request.Latitude, request.Longitude),
		request.Neighborhood, // ZonaID = Neighborhood
	)

	c.JSON(http.StatusCreated, response)
}

// DeleteTachoHandler elimina un tacho de MySQL y MongoDB
// @Summary Eliminar un tacho
// @Description Elimina un tacho (y sus características) de MySQL y MongoDB usando el ID de MySQL
// @Tags Tachos
// @Accept json
// @Produce json
// @Param id path int true "ID del tacho en MySQL"
// @Success 200 {object} map[string]string "Tacho eliminado exitosamente"
// @Failure 400 {object} map[string]string "Parámetros inválidos"
// @Failure 404 {object} map[string]string "Tacho no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /tachos/{id} [delete]
func DeleteTachoHandler(c *gin.Context) {
	// Obtener ID de MySQL del parámetro de ruta
	idStr := c.Param("id")

	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Debe proporcionar el 'id' del tacho",
		})
		return
	}

	var tachoID int
	if _, err := fmt.Sscanf(idStr, "%d", &tachoID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El ID debe ser un número válido"})
		return
	}

	// Eliminar el tacho usando el servicio
	err := services.DeleteTacho(tachoID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Tacho con ID %d no encontrado", tachoID)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 🔥 Publicar evento de eliminación a RabbitMQ
	go services.PublishTachoEliminado(tachoID, "Eliminado por usuario")

	c.JSON(http.StatusOK, gin.H{
		"message": "Tacho eliminado exitosamente (incluyendo características)",
		"id":      tachoID,
	})
}

// GetAllTachosHandler obtiene todos los tachos con información completa
// @Summary Obtener todos los tachos
// @Description Devuelve todos los tachos con neighborhood, latitud, longitud, estado y capacidad
// @Tags Tachos
// @Accept json
// @Produce json
// @Success 200 {array} services.TachoCompleto "Lista de todos los tachos"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /tachos [get]
func GetAllTachosHandler(c *gin.Context) {
	// Obtener todos los tachos usando el servicio
	tachos, err := services.GetAllTachos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tachos": tachos,
		"total":  len(tachos),
	})
}

// GetTachoByIDHandler obtiene un tacho específico por ID con todas sus características
// @Summary Obtener un tacho por ID
// @Description Devuelve un tacho específico con todas sus características, coordenadas y estado
// @Tags Tachos
// @Accept json
// @Produce json
// @Param id path int true "ID del tacho"
// @Success 200 {object} services.TachoCompleto "Tacho encontrado"
// @Failure 400 {object} map[string]string "ID inválido"
// @Failure 404 {object} map[string]string "Tacho no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /tachos/{id} [get]
func GetTachoByIDHandler(c *gin.Context) {
	// Obtener el ID del parámetro de la URL
	idStr := c.Param("id")

	var tachoID int
	if _, err := fmt.Sscanf(idStr, "%d", &tachoID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido, debe ser un número"})
		return
	}

	// Obtener el tacho por ID
	tacho, err := services.GetTachoByID(tachoID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Tacho con ID %d no encontrado", tachoID)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tacho)
}
