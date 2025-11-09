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

// CreateCentroHandler crea un nuevo centro en MySQL y MongoDB
// @Summary Crear un nuevo centro
// @Description Crea un centro con información de tipo (MySQL) y datos espaciales (MongoDB)
// @Tags Centros
// @Accept json
// @Produce json
// @Param centro body services.CreateCentroRequest true "Datos del centro a crear"
// @Success 201 {object} services.CreateCentroResponse "Centro creado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /centros [post]
func CreateCentroHandler(c *gin.Context) {
	var request services.CreateCentroRequest

	// Validar el JSON de entrada
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Datos inválidos: " + err.Error(),
		})
		return
	}

	// Llamar al servicio para crear el centro
	response, err := services.CreateCentro(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error al crear centro: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// DeleteCentroHandler elimina un centro de MySQL y MongoDB
// @Summary Eliminar un centro
// @Description Elimina un centro por su ID o por su ID de MongoDB
// @Tags Centros
// @Accept json
// @Produce json
// @Param id_centro query int false "ID del centro en MySQL"
// @Param id_mongo query int false "ID del centro en MongoDB"
// @Success 200 {object} map[string]string "Centro eliminado exitosamente"
// @Failure 400 {object} map[string]string "Parámetros inválidos"
// @Failure 404 {object} map[string]string "Centro no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /centros [delete]
func DeleteCentroHandler(c *gin.Context) {
	// Obtener parámetros de query
	idCentroStr := c.Query("id_centro")
	idMongoStr := c.Query("id_mongo")

	// Validar que se proporcione al menos uno
	if idCentroStr == "" && idMongoStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Debe proporcionar id_centro o id_mongo",
		})
		return
	}

	var centroID, mongoID int
	var err error

	// Convertir id_centro si se proporciona
	if idCentroStr != "" {
		centroID, err = strconv.Atoi(idCentroStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "id_centro debe ser un número entero",
			})
			return
		}
	}

	// Convertir id_mongo si se proporciona
	if idMongoStr != "" {
		mongoID, err = strconv.Atoi(idMongoStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "id_mongo debe ser un número entero",
			})
			return
		}
	}

	// Llamar al servicio para eliminar
	if err := services.DeleteCentro(centroID, mongoID); err != nil {
		if err.Error() == "no se encontró el centro para eliminar" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Centro no encontrado",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error al eliminar centro: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Centro eliminado exitosamente",
	})
}
