package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestBody define cómo esperamos los datos del formulario
type RequestBody struct {
	Tipo        string `json:"tipo" example:"incendio"`
	Descripcion string `json:"descripcion" example:"Incendio en edificio de oficinas"`
}

// ResponseBody define la respuesta de éxito
type ResponseBody struct {
	Message     string `json:"message" example:"Emergencia enviada correctamente"`
	Tipo        string `json:"tipo" example:"incendio"`
	Descripcion string `json:"descripcion" example:"Incendio en edificio de oficinas"`
}

// SendEmergencyHandler envía una emergencia
// @Summary Enviar emergencia
// @Description Registra una nueva emergencia en el sistema
// @Tags Emergencias
// @Accept json
// @Produce json
// @Param emergencia body RequestBody true "Datos de la emergencia"
// @Success 200 {object} ResponseBody "Emergencia enviada correctamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Router /enviar-emergencia [post]
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
