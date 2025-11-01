package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/gin-gonic/gin"
)

// PersonaWithEmailInfo contiene toda la información de una persona incluyendo su email
type PersonaWithEmailInfo struct {
	ID             string `json:"id"`
	Nombre         string `json:"nombre"`
	Email          string `json:"email"`
	NeighborhoodID int    `json:"neighborhood_id"`
	RouteNumber    int    `json:"route_number"`
	TruckID        int    `json:"truck_id"`
	RouteKey       string `json:"route_key"` // ej: "barrio_8_ruta_1"
}

// GetPersonasWithEmails obtiene todas las personas con sus emails para referencia
// @Summary Obtener listado de personas con emails
// @Description Devuelve todas las personas con sus emails asignados para saber qué email usar en cada ruta
// @Tags Personas
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Listado de personas con emails"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /personas/emails [get]
func GetPersonasWithEmails(c *gin.Context) {
	ctx := context.Background()

	personas, err := config.RedisClient.LRange(ctx, "personas", 0, -1).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo personas de Redis"})
		return
	}

	var result []PersonaWithEmailInfo

	for _, personKey := range personas {
		personData, err := config.RedisClient.HGetAll(ctx, personKey).Result()
		if err != nil {
			continue
		}

		neighborhoodID, _ := strconv.Atoi(personData["neighborhood_id"])
		routeNumber, _ := strconv.Atoi(personData["route_number"])
		truckID, _ := strconv.Atoi(personData["truck_id"])

		routeKey := "barrio_" + personData["neighborhood_id"] + "_ruta_" + personData["route_number"]

		persona := PersonaWithEmailInfo{
			ID:             personData["id"],
			Nombre:         personData["nombre"],
			Email:          personData["email"],
			NeighborhoodID: neighborhoodID,
			RouteNumber:    routeNumber,
			TruckID:        truckID,
			RouteKey:       routeKey,
		}

		result = append(result, persona)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    len(result),
		"message":  "Usa el campo 'email' para consultar /v2/ruta-optima-by-header",
		"personas": result,
	})
}
