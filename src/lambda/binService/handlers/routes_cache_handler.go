package handlers

import (
	"net/http"
	"strconv"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
)

// GetRoutesByNeighborhoodHandler obtiene todas las rutas de un barrio desde Redis
// @Summary Obtener rutas de un barrio
// @Description Devuelve todas las rutas simplificadas guardadas en caché para un barrio específico
// @Tags Rutas
// @Accept json
// @Produce json
// @Param neighborhood path int true "ID del barrio"
// @Success 200 {array} services.SimplifiedRoute "Lista de rutas del barrio"
// @Failure 400 {object} map[string]string "Neighborhood ID inválido"
// @Failure 404 {object} map[string]string "No se encontraron rutas"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /routes/neighborhood/{neighborhood} [get]
func GetRoutesByNeighborhoodHandler(c *gin.Context) {
	neighborhoodStr := c.Param("neighborhood")
	neighborhood, err := strconv.Atoi(neighborhoodStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Neighborhood ID inválido"})
		return
	}

	routes, err := services.GetAllRoutesForNeighborhood(neighborhood)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(routes) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No se encontraron rutas para este barrio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"neighborhood": neighborhood,
		"total_routes": len(routes),
		"routes":       routes,
	})
}

// GetSpecificRouteHandler obtiene una ruta específica desde Redis
// @Summary Obtener ruta específica
// @Description Devuelve una ruta simplificada específica desde caché
// @Tags Rutas
// @Accept json
// @Produce json
// @Param neighborhood path int true "ID del barrio"
// @Param routeNumber path int true "Número de ruta"
// @Success 200 {object} services.SimplifiedRoute "Ruta solicitada"
// @Failure 400 {object} map[string]string "Parámetros inválidos"
// @Failure 404 {object} map[string]string "Ruta no encontrada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /routes/neighborhood/{neighborhood}/route/{routeNumber} [get]
func GetSpecificRouteHandler(c *gin.Context) {
	neighborhoodStr := c.Param("neighborhood")
	routeNumberStr := c.Param("routeNumber")

	neighborhood, err := strconv.Atoi(neighborhoodStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Neighborhood ID inválido"})
		return
	}

	routeNumber, err := strconv.Atoi(routeNumberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Route number inválido"})
		return
	}

	route, err := services.GetSimplifiedRoute(neighborhood, routeNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruta no encontrada"})
		return
	}

	c.JSON(http.StatusOK, route)
}
