package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/middleware"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
)

var ctx = context.Background()

// PersonaResponse represents the response for persona data
type PersonaResponse struct {
	ID             string `json:"id"`
	Nombre         string `json:"nombre"`
	Email          string `json:"email"` // ✨ Nuevo: email asociado
	NeighborhoodID int    `json:"neighborhood_id"`
	RouteNumber    int    `json:"route_number"`
	TruckID        int    `json:"truck_id"`
}

// GetAllPersonas obtiene todas las personas de Redis
// @Summary Obtener todas las personas
// @Description Devuelve la lista completa de personas con sus asignaciones
// @Tags Personas
// @Accept json
// @Produce json
// @Success 200 {array} PersonaResponse "Lista de personas"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /personas [get]
func GetAllPersonas(c *gin.Context) {
	personas, err := config.RedisClient.LRange(ctx, "personas", 0, -1).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo personas de Redis"})
		return
	}

	var result []PersonaResponse

	// Obtener datos de cada persona
	for _, personKey := range personas {
		personData, err := config.RedisClient.HGetAll(ctx, personKey).Result()
		if err != nil {
			continue
		}

		neighborhoodID, _ := strconv.Atoi(personData["neighborhood_id"])
		routeNumber, _ := strconv.Atoi(personData["route_number"])
		truckID, _ := strconv.Atoi(personData["truck_id"])

		persona := PersonaResponse{
			ID:             personData["id"],
			Nombre:         personData["nombre"],
			Email:          personData["email"],
			NeighborhoodID: neighborhoodID,
			RouteNumber:    routeNumber,
			TruckID:        truckID,
		}

		result = append(result, persona)
	}

	// Actualizar métricas de Prometheus
	middleware.UpdatePersonasMetrics(len(result))

	// Contar personas por neighborhood para actualizar métricas
	neighborhoodCount := make(map[string]int)
	for _, persona := range result {
		key := strconv.Itoa(persona.NeighborhoodID)
		neighborhoodCount[key]++
	}
	for neighborhood, count := range neighborhoodCount {
		middleware.UpdatePersonasPorZona(neighborhood, count)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    len(result),
		"personas": result,
	})
}

// GetPersonaByID obtiene una persona específica por ID
// @Summary Obtener persona por ID
// @Description Devuelve los datos de una persona específica
// @Tags Personas
// @Accept json
// @Produce json
// @Param id path int true "ID de la persona"
// @Success 200 {object} PersonaResponse "Datos de la persona"
// @Failure 404 {object} map[string]string "Persona no encontrada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /personas/{id} [get]
func GetPersonaByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	personKey := "persona:" + strconv.Itoa(id)

	// Verificar si existe
	exists, err := config.RedisClient.Exists(ctx, personKey).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verificando existencia"})
		return
	}

	if exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Persona no encontrada"})
		return
	}

	// Obtener datos
	personData, err := config.RedisClient.HGetAll(ctx, personKey).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo datos"})
		return
	}

	neighborhoodID, _ := strconv.Atoi(personData["neighborhood_id"])
	routeNumber, _ := strconv.Atoi(personData["route_number"])
	truckID, _ := strconv.Atoi(personData["truck_id"])

	persona := PersonaResponse{
		ID:             personData["id"],
		Nombre:         personData["nombre"],
		Email:          personData["email"],
		NeighborhoodID: neighborhoodID,
		RouteNumber:    routeNumber,
		TruckID:        truckID,
	}

	c.JSON(http.StatusOK, persona)
}

// GetPersonasByNeighborhood obtiene personas de un barrio específico
// @Summary Obtener personas por barrio
// @Description Devuelve todas las personas asignadas a un barrio específico
// @Tags Personas
// @Accept json
// @Produce json
// @Param neighborhood path int true "Número de barrio"
// @Success 200 {array} PersonaResponse "Personas del barrio"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /personas/neighborhood/{neighborhood} [get]
func GetPersonasByNeighborhood(c *gin.Context) {
	neighborhoodStr := c.Param("neighborhood")
	neighborhood, err := strconv.Atoi(neighborhoodStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Neighborhood inválido"})
		return
	}

	// Obtener todas las personas
	personas, err := config.RedisClient.LRange(ctx, "personas", 0, -1).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo personas"})
		return
	}

	var result []PersonaResponse

	// Filtrar por neighborhood
	for _, personKey := range personas {
		personData, err := config.RedisClient.HGetAll(ctx, personKey).Result()
		if err != nil {
			continue
		}

		personNeighborhood, _ := strconv.Atoi(personData["neighborhood_id"])
		if personNeighborhood == neighborhood {
			routeNumber, _ := strconv.Atoi(personData["route_number"])
			truckID, _ := strconv.Atoi(personData["truck_id"])

			persona := PersonaResponse{
				ID:             personData["id"],
				Nombre:         personData["nombre"],
				Email:          personData["email"],
				NeighborhoodID: personNeighborhood,
				RouteNumber:    routeNumber,
				TruckID:        truckID,
			}
			result = append(result, persona)
		}
	}

	// Actualizar métricas de Prometheus para este barrio específico
	middleware.UpdatePersonasPorZona(neighborhoodStr, len(result))

	c.JSON(http.StatusOK, gin.H{
		"neighborhood": neighborhood,
		"total":        len(result),
		"personas":     result,
	})
}

// CreatePersonaHandler crea una nueva persona en Redis
// @Summary Crear persona
// @Description Crea una persona (hash persona:<id>, lista personas, mapeo email->id)
// @Tags Personas
// @Accept json
// @Produce json
// @Param persona body services.CreatePersonaRequest true "Datos de la persona"
// @Success 201 {object} services.Persona "Persona creada"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 500 {object} map[string]string "Error interno"
// @Router /personas [post]
func CreatePersonaHandler(c *gin.Context) {
	var req services.CreatePersonaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	persona, err := services.CreatePersonaRedis(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, persona)
}

// DeletePersonaHandler elimina una persona en Redis
// @Summary Eliminar persona
// @Description Borra el hash y mapeo email->id de una persona
// @Tags Personas
// @Produce json
// @Param id path int true "ID de la persona"
// @Success 200 {object} map[string]string "Persona eliminada"
// @Failure 400 {object} map[string]string "ID inválido"
// @Failure 404 {object} map[string]string "Persona no encontrada"
// @Failure 500 {object} map[string]string "Error interno"
// @Router /personas/{id} [delete]
func DeletePersonaHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := services.DeletePersonaRedis(id); err != nil {
		if err.Error() == fmt.Sprintf("persona %d not found", id) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Persona no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Persona eliminada", "id": id})
}
