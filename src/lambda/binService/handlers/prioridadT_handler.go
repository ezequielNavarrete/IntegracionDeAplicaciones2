package handlers

import (
	"net/http"
	"strconv"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Request para actualizar prioridad
type UpdatePrioridadRequest struct {
	Prioridad int `json:"prioridad" example:"2"`
}

type UpdatePrioridadResponse struct {
	Message   string `json:"message"`
	IDTacho   int64  `json:"id_tacho"`
	IDNeo     string `json:"id_neo"`
	Prioridad int    `json:"prioridad"`
}

// UpdatePrioridadTachoHandler actualiza la prioridad de un tacho
// @Summary Actualizar prioridad del tacho
// @Description Actualiza el campo prioridad de un tacho en Neo4j
// @Tags Tachos
// @Accept json
// @Produce json
// @Param id_tacho path int true "ID del tacho"
// @Param prioridad body UpdatePrioridadRequest true "Nueva prioridad del tacho"
// @Success 200 {object} UpdatePrioridadResponse "Prioridad actualizada correctamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 500 {object} map[string]string "Error interno"
// @Router /tachos/{id_tacho}/prioridad [put]
func UpdatePrioridadTachoHandler(c *gin.Context) {
	// ID en URL
	idStr := c.Param("id_tacho")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Body
	var body UpdatePrioridadRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	// Obtener id_neo desde MySQL/GORM (usa config.DB global)
	var tacho models.Tacho
	if err := config.DB.WithContext(c).First(&tacho, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tacho no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar tacho"})
		return
	}

	// Obtener sesión (NO cerrar driver). GetNeo4jSession() devuelve una session con contexto
	session, err := config.GetNeo4jSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al conectar con Neo4j"})
		return
	}
	// cerrar sesión al terminar la request
	defer session.Close(c.Request.Context())

	// Ejecutar update usando el contexto de la request
	ctx := c.Request.Context()
	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (t:Tacho {id: $id})
			SET t.prioridad = $prioridad
			RETURN t.id AS id
		`
		// usar tx.Run con el mismo ctx
		_, err := tx.Run(ctx, query, map[string]any{
			"id":        tacho.IDNeo,
			"prioridad": body.Prioridad,
		})
		return nil, err
	})
	if err != nil {
		// log opcional: log.Printf("neo update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar prioridad en Neo4j"})
		return
	}

	c.JSON(http.StatusOK, UpdatePrioridadResponse{
		Message:   "Prioridad actualizada correctamente",
		IDTacho:   tacho.IDTacho,
		IDNeo:     tacho.IDNeo,
		Prioridad: body.Prioridad,
	})
}
