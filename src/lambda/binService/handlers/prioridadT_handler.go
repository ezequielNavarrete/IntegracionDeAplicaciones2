package handlers

import (
	"net/http"
	"strconv"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
)

// UpdatePrioridadTachoHandler actualiza las características de un tacho
// @Summary Actualizar características del tacho
// @Description Actualiza una o más características de un tacho con nuevas prioridades (0=Nulo, 1=Bajo, 2=Medio, 3=Alto, 4=Urgente)
// @Tags Tachos
// @Accept json
// @Produce json
// @Param id_tacho path int true "ID del tacho"
// @Param caracteristicas body services.UpdateTachoCaracteristicasRequest true "Características a actualizar"
// @Success 200 {object} map[string]interface{} "Características actualizadas correctamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 404 {object} map[string]string "Tacho no encontrado"
// @Failure 500 {object} map[string]string "Error interno"
// @Router /tachos/{id_tacho}/prioridad [put]
func UpdatePrioridadTachoHandler(c *gin.Context) {
	// ID en URL
	idStr := c.Param("id_tacho")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Body
	var request services.UpdateTachoCaracteristicasRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	// Validar que haya al menos una característica
	if len(request.Caracteristicas) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe proporcionar al menos una característica"})
		return
	}

	// Actualizar características
	if err := services.UpdateTachoCaracteristicas(id, request); err != nil {
		if err.Error() == "tacho not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tacho no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                 "Características actualizadas correctamente",
		"id_tacho":                id,
		"caracteristicas_updated": len(request.Caracteristicas),
	})
}
