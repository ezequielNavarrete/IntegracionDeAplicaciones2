package handlers

import (
	"net/http"
	"strconv"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/middleware"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/models"
	"github.com/gin-gonic/gin"
)

// Request para actualizar capacidad
type UpdateCapacidadRequest struct {
	Capacidad float64 `json:"capacidad" example:"90"`
}

// Response después de actualizar
type UpdateCapacidadResponse struct {
	Message   string  `json:"message" example:"Capacidad actualizada correctamente"`
	IDTacho   int64   `json:"id_tacho" example:"1"`
	Capacidad float64 `json:"capacidad" example:"90"`
}

// UpdateCapacidadHandler actualiza la capacidad de un tacho
// @Summary Actualizar capacidad del tacho
// @Description Actualiza el campo capacidad de un tacho en MySQL
// @Tags Tachos
// @Accept json
// @Produce json
// @Param id path int true "ID del tacho"
// @Param capacidad body UpdateCapacidadRequest true "Nueva capacidad del tacho"
// @Success 200 {object} UpdateCapacidadResponse "Capacidad actualizada correctamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 500 {object} map[string]string "Error interno"
// @Router /tachos/{id}/capacidad [put]
func UpdateCapacidadTachoHandler(c *gin.Context) {
	idStr := c.Param("id_tacho")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var body UpdateCapacidadRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	// >>> Validación de capacidad entre 0 y 100 <<<
	if body.Capacidad < 0 || body.Capacidad > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Capacidad fuera de rango"})
		return
	}

	if err := config.DB.Model(&models.Tacho{}).Where("id_tacho = ?", id).Update("capacidad", body.Capacidad).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar capacidad"})
		return
	}

	// Actualizar métricas de Prometheus
	// Nota: Necesitarías obtener la zona del tacho para las etiquetas completas
	middleware.UpdateTachoCapacidad(idStr, "zona_desconocida", body.Capacidad)

	c.JSON(http.StatusOK, UpdateCapacidadResponse{
		Message:   "Capacidad actualizada correctamente",
		IDTacho:   int64(id),
		Capacidad: body.Capacidad,
	})
}
