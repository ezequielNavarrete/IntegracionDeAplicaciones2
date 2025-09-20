package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestBody define cómo esperamos los datos del formulario
type RequestBody struct {
	Tipo        string `json:"tipo"`
	Descripcion string `json:"descripcion"`
}

// Handler simulates successfully uploading an emergency
func SendEmergencyHandler(c *gin.Context) {
	var body RequestBody

	// BindJSON intenta convertir el JSON enviado a la estructura RequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Emergencia enviada correctamente",
		"tipo":        body.Tipo,
		"descripcion": body.Descripcion,
	})

}
