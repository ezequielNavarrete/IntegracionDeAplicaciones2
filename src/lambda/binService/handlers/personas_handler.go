package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/gin-gonic/gin"
)

var ctx = context.Background()

// PersonaResponse represents the response for persona data
type PersonaResponse struct {
	ID       string `json:"id"`
	ZonaID   string `json:"zona_id"`
	CamionID string `json:"camion_id"`
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
	config.InitializePersonsData()
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

		persona := PersonaResponse{
			ID:       personData["id"],
			ZonaID:   personData["zona_id"],
			CamionID: personData["camion_id"],
		}

		result = append(result, persona)
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

	config.InitializePersonsData()
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

	persona := PersonaResponse{
		ID:       personData["id"],
		ZonaID:   personData["zona_id"],
		CamionID: personData["camion_id"],
	}

	c.JSON(http.StatusOK, persona)
}

// GetPersonasByZona obtiene personas de una zona específica
// @Summary Obtener personas por zona
// @Description Devuelve todas las personas asignadas a una zona específica
// @Tags Personas
// @Accept json
// @Produce json
// @Param zona path int true "Número de zona"
// @Success 200 {array} PersonaResponse "Personas de la zona"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /personas/zona/{zona} [get]
func GetPersonasByZona(c *gin.Context) {
	zonaStr := c.Param("zona")
	zona, err := strconv.Atoi(zonaStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zona inválida"})
		return
	}

	config.InitializePersonsData()
	// Obtener todas las personas
	personas, err := config.RedisClient.LRange(ctx, "personas", 0, -1).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo personas"})
		return
	}

	var result []PersonaResponse

	// Filtrar por zona
	for _, personKey := range personas {
		personData, err := config.RedisClient.HGetAll(ctx, personKey).Result()
		if err != nil {
			continue
		}

		personZona, _ := strconv.Atoi(personData["zona_id"])
		if personZona == zona {
			persona := PersonaResponse{
				ID:       personData["id"],
				ZonaID:   personData["zona_id"],
				CamionID: personData["camion_id"],
			}
			result = append(result, persona)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"zona":     zona,
		"total":    len(result),
		"personas": result,
	})
}
