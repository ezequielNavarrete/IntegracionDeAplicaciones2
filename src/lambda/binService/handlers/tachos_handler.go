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

	c.JSON(http.StatusCreated, response)
}

// DeleteTachoHandler elimina un tacho de MySQL y MongoDB
// @Summary Eliminar un tacho
// @Description Elimina un tacho tanto de MySQL como de MongoDB usando el ID de MongoDB
// @Tags Tachos
// @Accept json
// @Produce json
// @Param id query int true "ID del tacho en MongoDB"
// @Success 200 {object} map[string]string "Tacho eliminado exitosamente"
// @Failure 400 {object} map[string]string "Parámetros inválidos"
// @Failure 404 {object} map[string]string "Tacho no encontrado en ninguna base"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /tachos [delete]
func DeleteTachoHandler(c *gin.Context) {
	// Obtener ID de MongoDB del query parameter
	mongoIDStr := c.Query("id")

	if mongoIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Debe proporcionar el 'id' del tacho en MongoDB",
			"example": "?id=50006",
		})
		return
	}

	var mongoID int
	if _, err := fmt.Sscanf(mongoIDStr, "%d", &mongoID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El ID debe ser un número válido"})
		return
	}

	// Eliminar el tacho usando el servicio
	err := services.DeleteTacho(mongoID)
	if err != nil {
		if strings.Contains(err.Error(), "no se pudo eliminar de ninguna base de datos") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tacho no encontrado en ninguna base de datos"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tacho eliminado exitosamente",
		"id":      mongoID,
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
