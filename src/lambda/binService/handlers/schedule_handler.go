package handlers

import (
	"net/http"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
)

// GetScheduleConfigHandler obtiene la configuración actual de horarios
// @Summary Obtener configuración de horarios
// @Description Devuelve la configuración actual del cron, próxima recolección y última actualización
// @Tags Schedule
// @Accept json
// @Produce json
// @Success 200 {object} services.ScheduleConfig "Configuración de horarios"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /schedule [get]
func GetScheduleConfigHandler(c *gin.Context) {
	config, err := services.GetScheduleConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateScheduleConfigHandler actualiza la configuración del horario del cron
// @Summary Actualizar configuración de horarios
// @Description Actualiza el horario del cron de recolección (formato cron: "minuto hora día mes día_semana"). Opcionalmente puede incluir una descripción personalizada.
// @Tags Schedule
// @Accept json
// @Produce json
// @Param schedule body services.UpdateScheduleRequest true "Nueva configuración de cron"
// @Success 200 {object} services.ScheduleConfig "Configuración actualizada"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /schedule [put]
func UpdateScheduleConfigHandler(c *gin.Context) {
	var request services.UpdateScheduleRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Datos inválidos: " + err.Error(),
			"example": map[string]interface{}{
				"cron_schedule": "0 16 * * *",
				"description":   "Recolección diaria por la tarde",
			},
			"help": "Formato cron: 'minuto hora día mes día_semana'. Ejemplos: '0 16 * * *' (4 PM diario), '0 */6 * * *' (cada 6 horas). El campo 'description' es opcional.",
		})
		return
	}

	config, err := services.UpdateScheduleConfig(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuración de horario actualizada exitosamente",
		"config":  config,
		"note":    "Recuerda actualizar tu cron job en Render o donde esté configurado",
	})
}
