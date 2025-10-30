package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
)

// GetRutaHandlerV2 obtiene la ruta asignada al usuario autenticado
// @Summary Obtener ruta óptima del usuario (v2)
// @Description Devuelve la ruta óptima asignada al usuario basándose en el email del JWT
// @Tags Rutas V2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.SimplifiedRoute "Ruta asignada al usuario"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Usuario no encontrado o sin ruta asignada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /v2/ruta-optima [get]
func GetRutaHandlerV2(c *gin.Context) {
	// Obtener email del contexto (ya fue validado por el middleware JWT)
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No se pudo obtener el email del token"})
		return
	}

	emailStr, ok := email.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email inválido en token"})
		return
	}

	// Obtener el user_id desde Redis usando el email
	ctx := context.Background()
	userID, err := config.RedisClient.Get(ctx, emailStr).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// Buscar la persona asociada a este user_id
	personKey := "persona:" + userID
	existsCount, err := config.RedisClient.Exists(ctx, personKey).Result()
	if err != nil || existsCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No se encontró persona asociada al usuario"})
		return
	}

	// Obtener datos de la persona
	personData, err := config.RedisClient.HGetAll(ctx, personKey).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo datos de la persona"})
		return
	}

	// Extraer neighborhood_id y route_number
	neighborhoodID, err := strconv.Atoi(personData["neighborhood_id"])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Neighborhood ID inválido"})
		return
	}

	routeNumber, err := strconv.Atoi(personData["route_number"])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Route number inválido"})
		return
	}

	// Obtener la ruta desde Redis
	route, err := services.GetSimplifiedRoute(neighborhoodID, routeNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruta no encontrada en caché"})
		return
	}

	// Agregar información del usuario a la respuesta
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"email":        emailStr,
			"persona_id":   personData["id"],
			"nombre":       personData["nombre"],
			"neighborhood": neighborhoodID,
			"route_number": routeNumber,
		},
		"route": route,
	})
}

// GetRutaHandlerV2ByHeader es similar a GetRutaHandlerV2 pero obtiene el email del header
// @Summary Obtener ruta óptima del usuario por header (v2)
// @Description Devuelve la ruta óptima asignada al usuario basándose en el header X-User-Email
// @Tags Rutas V2
// @Accept json
// @Produce json
// @Param X-User-Email header string true "Email del usuario"
// @Success 200 {object} services.SimplifiedRoute "Ruta asignada al usuario"
// @Failure 400 {object} map[string]string "Header faltante"
// @Failure 404 {object} map[string]string "Usuario no encontrado o sin ruta asignada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /v2/ruta-optima-by-header [get]
func GetRutaHandlerV2ByHeader(c *gin.Context) {
	// Obtener email del header
	emailStr := c.GetHeader("X-User-Email")
	if emailStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Header X-User-Email requerido"})
		return
	}

	// Obtener el user_id desde Redis usando el email
	ctx := context.Background()
	userID, err := config.RedisClient.Get(ctx, emailStr).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// Buscar la persona asociada a este user_id
	personKey := "persona:" + userID
	existsCount, err := config.RedisClient.Exists(ctx, personKey).Result()
	if err != nil || existsCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No se encontró persona asociada al usuario"})
		return
	}

	// Obtener datos de la persona
	personData, err := config.RedisClient.HGetAll(ctx, personKey).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo datos de la persona"})
		return
	}

	// Extraer neighborhood_id y route_number
	neighborhoodID, err := strconv.Atoi(personData["neighborhood_id"])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Neighborhood ID inválido"})
		return
	}

	routeNumber, err := strconv.Atoi(personData["route_number"])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Route number inválido"})
		return
	}

	// Obtener la ruta desde Redis
	route, err := services.GetSimplifiedRoute(neighborhoodID, routeNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruta no encontrada en caché"})
		return
	}

	// Agregar información del usuario a la respuesta
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"email":        emailStr,
			"persona_id":   personData["id"],
			"nombre":       personData["nombre"],
			"neighborhood": neighborhoodID,
			"route_number": routeNumber,
		},
		"route": route,
	})
}
