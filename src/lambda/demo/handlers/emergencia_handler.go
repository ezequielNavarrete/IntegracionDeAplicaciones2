package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler simulates successfully uploading an emergency
func SendEmergencyHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Emergencia enviada"})
}
