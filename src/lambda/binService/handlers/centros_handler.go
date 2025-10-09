package handlers

import (
	"net/http"
	"strconv"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
)

// GetAllCentrosHandler obtiene todos los centros con información completa de MySQL y Neo4j
// @Summary Obtener todos los centros
// @Description Obtiene una lista de todos los centros con información de tipo (MySQL) y datos adicionales como nombre, barrio, dirección, coordenadas (Neo4j)
// @Tags Centros
// @Produce json
// @Success 200 {object} services.CentrosResponse "Lista de centros obtenida exitosamente"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /centros [get]
func GetAllCentrosHandler(c *gin.Context) {
	// Llamar al servicio para obtener todos los centros
	response, err := services.GetAllCentros()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error al obtener centros: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetCentroByIDHandler obtiene un centro específico por ID con información completa
// @Summary Obtener un centro por ID
// @Description Obtiene la información completa de un centro específico mediante su ID, incluyendo tipo (MySQL) y datos adicionales como nombre, barrio, dirección, coordenadas (Neo4j)
// @Tags Centros
// @Produce json
// @Param id path int true "ID del centro"
// @Success 200 {object} services.CentroResponse "Centro obtenido exitosamente"
// @Failure 400 {object} map[string]string "ID de centro inválido"
// @Failure 404 {object} map[string]string "Centro no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /centros/{id} [get]
func GetCentroByIDHandler(c *gin.Context) {
	// Obtener el ID del parámetro de la URL
	idParam := c.Param("id")

	// Convertir el ID a entero
	centroID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID de centro inválido: debe ser un número entero",
		})
		return
	}

	// Validar que el ID sea positivo
	if centroID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID de centro debe ser mayor a 0",
		})
		return
	}

	// Llamar al servicio para obtener el centro
	response, err := services.GetCentroByID(centroID)
	if err != nil {
		// Verificar si es un error de "no encontrado"
		if err.Error() == "centro with ID "+idParam+" not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Centro no encontrado con ID: " + idParam,
			})
			return
		}

		// Error interno del servidor (puede incluir errores de Neo4j)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error al obtener centro: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
