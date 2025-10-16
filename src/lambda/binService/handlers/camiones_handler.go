package handlers

import (
	"net/http"
	"strconv"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
)

// GetAllCamionesHandler obtiene todos los camiones con información de tipo y estado
// @Summary Obtener todos los camiones
// @Description Obtiene una lista de todos los camiones con información de tipo y estado mediante JOINs
// @Tags Camiones
// @Produce json
// @Success 200 {object} services.CamionesResponse "Lista de camiones obtenida exitosamente"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /camiones [get]
func GetAllCamionesHandler(c *gin.Context) {
	// Llamar al servicio para obtener todos los camiones
	response, err := services.GetAllCamiones()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error al obtener camiones: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetCamionByIDHandler obtiene un camión específico por ID
// @Summary Obtener un camión por ID
// @Description Obtiene la información completa de un camión específico mediante su ID, incluyendo tipo y estado
// @Tags Camiones
// @Produce json
// @Param id path int true "ID del camión"
// @Success 200 {object} services.CamionResponse "Camión obtenido exitosamente"
// @Failure 400 {object} map[string]string "ID de camión inválido"
// @Failure 404 {object} map[string]string "Camión no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /camiones/{id} [get]
func GetCamionByIDHandler(c *gin.Context) {
	// Obtener el ID del parámetro de la URL
	idParam := c.Param("id")

	// Convertir el ID a entero
	camionID, err := strconv.Atoi(idParam)
	if err != nil {
		// Log the error for debugging purposes
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID de camión inválido: debe ser un número entero",
		})
		return
	}

	// Validar que el ID sea positivo
	if camionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID de camión debe ser mayor a 0",
		})
		return
	}

	// Llamar al servicio para obtener el camión
	response, err := services.GetCamionByID(camionID)
	if err != nil {
		// Verificar si es un error de "no encontrado"
		if err.Error() == "camion with ID "+idParam+" not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Camión no encontrado con ID: " + idParam,
			})
			return
		}

		// Error interno del servidor
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error al obtener camión: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
