package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/events/schemas"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RouteNavigationResponse representa la respuesta de navegación punto a punto
type RouteNavigationResponse struct {
	RouteID      string               `json:"route_id"`
	CurrentIndex int                  `json:"current_index"`
	TotalPoints  int                  `json:"total_points"`
	CurrentPoint *services.Coordinate `json:"current_point"`
	NextPoint    *services.Coordinate `json:"next_point,omitempty"`
	IsFirstPoint bool                 `json:"is_first_point"`
	IsLastPoint  bool                 `json:"is_last_point"`
	Progress     float64              `json:"progress_percentage"`
	NextPointKey string               `json:"next_point_key,omitempty"`     // Clave para obtener el siguiente punto
	PreviousKey  string               `json:"previous_point_key,omitempty"` // Clave para volver al anterior
}

// GetRouteNavigationHandler obtiene el punto actual y el siguiente de una ruta
// @Summary Obtener navegación punto a punto de una ruta
// @Description Devuelve el punto actual y el siguiente punto en la ruta. Útil para navegación secuencial
// @Tags Rutas v2
// @Accept json
// @Produce json
// @Param neighborhood path int true "ID del barrio"
// @Param route path int true "Número de ruta"
// @Param index query int false "Índice del punto actual (0 = inicio)" default(0)
// @Success 200 {object} RouteNavigationResponse "Información de navegación"
// @Failure 400 {object} map[string]string "Parámetros inválidos"
// @Failure 404 {object} map[string]string "Ruta no encontrada"
// @Failure 500 {object} map[string]string "Error interno"
// @Router /v2/ruta-navegacion/{neighborhood}/{route} [get]
func GetRouteNavigationHandler(c *gin.Context) {
	// Obtener parámetros de la ruta
	neighborhoodStr := c.Param("neighborhood")
	routeStr := c.Param("route")
	indexStr := c.DefaultQuery("index", "0")

	neighborhood, err := strconv.Atoi(neighborhoodStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neighborhood debe ser un número válido"})
		return
	}

	routeNumber, err := strconv.Atoi(routeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "route debe ser un número válido"})
		return
	}

	currentIndex, err := strconv.Atoi(indexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index debe ser un número válido"})
		return
	}

	// Obtener la ruta desde Redis
	route, err := services.GetSimplifiedRoute(neighborhood, routeNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Ruta no encontrada: %v", err)})
		return
	}

	// Validar que el índice esté dentro del rango
	if currentIndex < 0 || currentIndex >= len(route.BinsCoords) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Índice fuera de rango. Debe estar entre 0 y %d", len(route.BinsCoords)-1),
		})
		return
	}

	// Construir la respuesta de navegación
	response := RouteNavigationResponse{
		RouteID:      route.RouteID,
		CurrentIndex: currentIndex,
		TotalPoints:  len(route.BinsCoords),
		CurrentPoint: &route.BinsCoords[currentIndex],
		IsFirstPoint: currentIndex == 0,
		IsLastPoint:  currentIndex == len(route.BinsCoords)-1,
		Progress:     float64(currentIndex) / float64(len(route.BinsCoords)-1) * 100,
	}

	// Agregar el siguiente punto si no es el último
	if currentIndex < len(route.BinsCoords)-1 {
		response.NextPoint = &route.BinsCoords[currentIndex+1]
		response.NextPointKey = fmt.Sprintf("/v2/ruta-navegacion/%d/%d?index=%d", neighborhood, routeNumber, currentIndex+1)
	}

	// Agregar clave para punto anterior si no es el primero
	if currentIndex > 0 {
		response.PreviousKey = fmt.Sprintf("/v2/ruta-navegacion/%d/%d?index=%d", neighborhood, routeNumber, currentIndex-1)
	}

	// Publicar evento de navegación (no bloquear si falla)
	eventosIDs := getEventosIDsFromCurrentPoint(response.CurrentPoint)
	if err := publishRutaNavegacionEvent(c, route.RouteID, currentIndex, len(route.BinsCoords), response.CurrentPoint, eventosIDs); err != nil {
		log.Printf("⚠️  Error publicando evento de navegación: %v", err)
	}

	c.JSON(http.StatusOK, response)
}

// GetRouteNavigationByHeaderHandler obtiene navegación punto a punto usando header de email
// @Summary Obtener navegación punto a punto usando email del conductor
// @Description Devuelve el punto actual y siguiente de la ruta asignada al conductor
// @Tags Rutas v2
// @Accept json
// @Produce json
// @Param X-User-Email header string true "Email del conductor"
// @Param index query int false "Índice del punto actual (0 = inicio)" default(0)
// @Success 200 {object} RouteNavigationResponse "Información de navegación"
// @Failure 400 {object} map[string]string "Parámetros inválidos"
// @Failure 404 {object} map[string]string "Usuario o ruta no encontrada"
// @Failure 500 {object} map[string]string "Error interno"
// @Router /v2/ruta-navegacion-by-header [get]
func GetRouteNavigationByHeaderHandler(c *gin.Context) {
	// Obtener email del header
	email := c.GetHeader("X-User-Email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Header X-User-Email es requerido"})
		return
	}

	indexStr := c.DefaultQuery("index", "0")
	currentIndex, err := strconv.Atoi(indexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index debe ser un número válido"})
		return
	}

	// Obtener userID desde Redis
	userID, err := services.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Usuario no encontrado: %v", err)})
		return
	}

	// Buscar la persona asociada al userID
	personaKey := fmt.Sprintf("persona:%s", userID)
	persona, err := services.GetPersonaByKey(personaKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Persona no encontrada: %v", err)})
		return
	}

	// Extraer neighborhood_id y route_number
	neighborhoodID, _ := strconv.Atoi(persona["neighborhood_id"].(string))
	routeNumber, _ := strconv.Atoi(persona["route_number"].(string))

	// Obtener la ruta desde Redis
	route, err := services.GetSimplifiedRoute(neighborhoodID, routeNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Ruta no encontrada: %v", err)})
		return
	}

	// Validar que el índice esté dentro del rango
	if currentIndex < 0 || currentIndex >= len(route.BinsCoords) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Índice fuera de rango. Debe estar entre 0 y %d", len(route.BinsCoords)-1),
		})
		return
	}

	// Construir la respuesta de navegación
	response := RouteNavigationResponse{
		RouteID:      route.RouteID,
		CurrentIndex: currentIndex,
		TotalPoints:  len(route.BinsCoords),
		CurrentPoint: &route.BinsCoords[currentIndex],
		IsFirstPoint: currentIndex == 0,
		IsLastPoint:  currentIndex == len(route.BinsCoords)-1,
		Progress:     float64(currentIndex) / float64(len(route.BinsCoords)-1) * 100,
	}

	// Agregar el siguiente punto si no es el último
	if currentIndex < len(route.BinsCoords)-1 {
		response.NextPoint = &route.BinsCoords[currentIndex+1]
		response.NextPointKey = fmt.Sprintf("/v2/ruta-navegacion-by-header?index=%d", currentIndex+1)
	}

	// Agregar clave para punto anterior si no es el primero
	if currentIndex > 0 {
		response.PreviousKey = fmt.Sprintf("/v2/ruta-navegacion-by-header?index=%d", currentIndex-1)
	}

	// Publicar evento de navegación (no bloquear si falla)
	eventosIDs := getEventosIDsFromCurrentPoint(response.CurrentPoint)
	if err := publishRutaNavegacionEvent(c, route.RouteID, currentIndex, len(route.BinsCoords), response.CurrentPoint, eventosIDs); err != nil {
		log.Printf("⚠️  Error publicando evento de navegación: %v", err)
	}

	c.JSON(http.StatusOK, response)
}

// GetRouteStartHandler obtiene el punto inicial de la ruta (index 0)
// @Summary Iniciar navegación de ruta
// @Description Obtiene el primer punto (depot) de la ruta para comenzar la navegación
// @Tags Rutas v2
// @Accept json
// @Produce json
// @Param neighborhood path int true "ID del barrio"
// @Param route path int true "Número de ruta"
// @Success 200 {object} RouteNavigationResponse "Punto inicial"
// @Failure 400 {object} map[string]string "Parámetros inválidos"
// @Failure 404 {object} map[string]string "Ruta no encontrada"
// @Router /v2/ruta-navegacion/{neighborhood}/{route}/start [get]
func GetRouteStartHandler(c *gin.Context) {
	// Redirigir a la navegación con index 0
	c.Request.URL.RawQuery = "index=0"
	GetRouteNavigationHandler(c)
}

// GetRouteStartByHeaderHandler obtiene el punto inicial usando email
// @Summary Iniciar navegación usando email del conductor
// @Description Obtiene el primer punto de la ruta asignada al conductor
// @Tags Rutas v2
// @Accept json
// @Produce json
// @Param X-User-Email header string true "Email del conductor"
// @Success 200 {object} RouteNavigationResponse "Punto inicial"
// @Failure 400 {object} map[string]string "Header faltante"
// @Failure 404 {object} map[string]string "Usuario o ruta no encontrada"
// @Router /v2/ruta-navegacion-by-header/start [get]
func GetRouteStartByHeaderHandler(c *gin.Context) {
	// Redirigir a la navegación con index 0
	c.Request.URL.RawQuery = "index=0"
	GetRouteNavigationByHeaderHandler(c)
}

// publishRutaNavegacionEvent publica el evento de progreso de navegación
func publishRutaNavegacionEvent(c *gin.Context, routeID string, currentIndex, totalPoints int, currentPoint *services.Coordinate, eventosIDs []string) error {
	ctx := c.Request.Context()

	// Hardcodear IDs de eventos culturales para testing
	eventosIDsHardcoded := []string{"691f937a6fab2549f2258472", "68d44d2663d135b1b22cb971"}

	// Crear array de información adicional con los IDs de eventos hardcodeados
	informacionAdicional := make([]schemas.InformacionEvento, 0, len(eventosIDsHardcoded))
	for _, idEvento := range eventosIDsHardcoded {
		informacionAdicional = append(informacionAdicional, schemas.InformacionEvento{
			IDEvento: idEvento,
		})
	}

	// Calcular porcentaje de progreso
	porcentajeProgreso := 0.0
	if totalPoints > 1 {
		porcentajeProgreso = float64(currentIndex) / float64(totalPoints-1) * 100
	}

	// Crear payload
	payload := schemas.RutaNavegacionPayload{
		IDRuta:            routeID,
		IndicePuntoActual: currentIndex,
		TotalPuntos:       totalPoints,
		PuntoActual: schemas.PuntoNavegacion{
			Latitud:  currentPoint.Lat,
			Longitud: currentPoint.Lon,
		},
		PorcentajeProgreso:   porcentajeProgreso,
		InformacionAdicional: informacionAdicional,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Crear envelope estándar
	envelope := schemas.EventEnvelope{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Source:    "residuos",
		Topic:     schemas.RoutingKeyRutaNavegacion,
		Payload:   payloadBytes,
	}

	// Publicar evento
	if err := events.Publish(ctx, schemas.RoutingKeyRutaNavegacion, envelope); err != nil {
		return err
	}

	log.Printf("✅ [RutaNavegacion] Evento publicado - Routing Key: %s", schemas.RoutingKeyRutaNavegacion)
	log.Printf("   📍 Ruta: %s | Punto: %d/%d (%.2f%%)", routeID, currentIndex+1, totalPoints, porcentajeProgreso)
	if len(eventosIDs) > 0 {
		log.Printf("   🎭 Eventos culturales en este punto: %v", eventosIDs)
	}

	return nil
}

// getEventosIDsFromCurrentPoint obtiene los IDs de eventos culturales del punto actual
// buscando en MongoDB los tachos en esas coordenadas que tengan id_evento
func getEventosIDsFromCurrentPoint(point *services.Coordinate) []string {
	if point == nil {
		return []string{}
	}

	// Obtener colección de MongoDB
	mongoCollection, err := config.GetMongoCollection("tachos")
	if err != nil {
		log.Printf("Error obteniendo colección MongoDB: %v", err)
		return []string{}
	}

	ctx := context.Background()

	// Buscar tachos en las coordenadas exactas que tengan id_evento
	filter := map[string]interface{}{
		"lat": point.Lat,
		"lon": point.Lon,
		"id_evento": map[string]interface{}{
			"$exists": true,
			"$ne":     "",
		},
	}

	cursor, err := mongoCollection.Find(ctx, filter)
	if err != nil {
		log.Printf("Error buscando tachos con eventos: %v", err)
		return []string{}
	}
	defer cursor.Close(ctx)

	// Extraer los IDs de eventos
	eventosIDs := []string{}
	for cursor.Next(ctx) {
		var tacho map[string]interface{}
		if err := cursor.Decode(&tacho); err != nil {
			log.Printf("Error decodificando tacho: %v", err)
			continue
		}

		if idEvento, ok := tacho["id_evento"].(string); ok && idEvento != "" {
			eventosIDs = append(eventosIDs, idEvento)
		}
	}

	return eventosIDs
}
