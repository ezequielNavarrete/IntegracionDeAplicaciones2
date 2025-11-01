package handlers

import (
	"net/http"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
)

// RegeneratePersonasHandler regenera todas las personas desde las rutas cacheadas
// @Summary Regenerar personas desde rutas
// @Description Regenera todas las personas con sus emails desde las rutas existentes en cache
// @Tags Personas
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Personas regeneradas exitosamente"
// @Failure 500 {object} map[string]string "Error regenerando personas"
// @Router /personas/regenerate [post]
func RegeneratePersonasHandler(c *gin.Context) {
	if err := services.GeneratePersonasFromRoutes(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error regenerando personas",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Personas regeneradas exitosamente con emails",
		"info":    "Consulta GET /personas/emails para ver todos los conductores",
	})
}
