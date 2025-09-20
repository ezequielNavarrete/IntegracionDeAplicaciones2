package handlers

import (
	"net/http"
	"strconv"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
)

// GetRutaHandler obtiene la ruta óptima para una zona específica
// @Summary Obtener ruta óptima
// @Description Devuelve la ruta óptima y distancias para una zona específica
// @Tags Rutas
// @Accept json
// @Produce json
// @Param zonaID path int true "ID de la zona"
// @Success 200 {array} map[string]interface{} "Lista de puntos con distancias"
// @Failure 400 {object} map[string]string "zonaID inválido"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /ruta-optima/{zonaID} [get]
func GetRutaHandler(c *gin.Context) {
	zonaIDStr := c.Param("zonaID")
	zonaID, err := strconv.Atoi(zonaIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "zonaID inválido"})
		return
	}

	points, err := services.GetDistances(zonaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, points)
}
