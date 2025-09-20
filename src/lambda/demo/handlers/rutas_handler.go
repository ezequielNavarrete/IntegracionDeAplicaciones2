package handlers

import (
	"net/http"
	"strconv"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/demo/services"
	"github.com/gin-gonic/gin"
)

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
