package handlers

import (
	"net/http"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
)

// GetScheduleHandler obtiene la configuración global de horarios
// @Summary Obtener configuración de horarios de recolección
// @Description Devuelve los horarios de recolección por neighborhood y las fechas de actualización de rutas
// @Tags Schedule
// @Produce json
// @Success 200 {object} services.GlobalScheduleConfig
// @Router /schedule [get]
func GetScheduleHandler(c *gin.Context) {
	config, err := services.GetGlobalSchedule()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateCronScheduleHandler actualiza el horario del cron (cuándo se calculan rutas)
// @Summary Actualizar horario del cron
// @Description Actualiza cuándo se ejecuta el cron que calcula las rutas
// @Tags Schedule
// @Accept json
// @Produce json
// @Param request body services.UpdateScheduleRequest true "Cron schedule"
// @Success 200 {object} services.GlobalScheduleConfig
// @Router /schedule/cron [put]
func UpdateCronScheduleHandler(c *gin.Context) {
	var req services.UpdateScheduleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de request inválido", "details": err.Error()})
		return
	}

	config, err := services.UpdateCronSchedule(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Horario del cron actualizado exitosamente",
		"config":  config,
	})
}

// UpdateNeighborhoodScheduleHandler actualiza el horario de recolección de un neighborhood
// @Summary Actualizar horario de recolección de un neighborhood
// @Description Actualiza las horas de inicio y fin de recolección para un neighborhood específico
// @Tags Schedule
// @Accept json
// @Produce json
// @Param request body services.UpdateNeighborhoodScheduleRequest true "Neighborhood schedule"
// @Success 200 {object} services.GlobalScheduleConfig
// @Router /schedule/neighborhood [put]
func UpdateNeighborhoodScheduleHandler(c *gin.Context) {
	var req services.UpdateNeighborhoodScheduleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de request inválido", "details": err.Error()})
		return
	}

	config, err := services.UpdateNeighborhoodSchedule(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Horario del neighborhood actualizado exitosamente",
		"config":  config,
	})
}
