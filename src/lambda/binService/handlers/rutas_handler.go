package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/middleware"
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
	start := time.Now() // Para medir tiempo de cálculo

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

	// Registrar métricas de negocio
	duration := time.Since(start).Seconds()
	middleware.IncrementRutasOptimas(zonaIDStr)
	middleware.ObserveRutaCalculoTime(zonaIDStr, duration)

	c.JSON(http.StatusOK, points)
}

// GetRutaHandlerByHeader obtiene la ruta óptima basada en el email del header
// @Summary Obtener ruta óptima por email
// @Description Devuelve la ruta óptima y distancias para la zona de la persona asociada al email
// @Tags Rutas
// @Accept json
// @Produce json
// @Param email header string true "email del usuario"
// @Success 200 {array} map[string]interface{} "Lista de puntos con distancias"
// @Failure 400 {object} map[string]string "Email faltante o inválido"
// @Failure 404 {object} map[string]string "Usuario o persona no encontrada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /ruta-optima [get]
func GetRutaHandlerByHeader(c *gin.Context) {
	// Obtener email del header
	email := c.GetHeader("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Header 'Email' es requerido"})
		return
	}

	// Consultar Redis para obtener el número de persona por email
	personaNumStr, err := services.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado: " + err.Error()})
		return
	}

	// Buscar la persona en Redis usando persona:x
	personaKey := "persona:" + personaNumStr
	persona, err := services.GetPersonaByKey(personaKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Persona no encontrada: " + err.Error()})
		return
	}

	// Convertir zona_id a entero
	zonaIDStr, ok := persona["zona_id"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "zona_id inválida en persona"})
		return
	}

	zonaID, err := strconv.Atoi(zonaIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "zona_id no es un número válido"})
		return
	}

	// Obtener las distancias/rutas para la zona
	points, err := services.GetDistances(zonaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email":     email,
		"persona":   personaNumStr,
		"zona_id":   zonaID,
		"zona_name": persona["zona_nombre"],
		"routes":    points,
	})
}
