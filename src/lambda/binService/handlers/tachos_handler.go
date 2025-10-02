package handlers

import (
	"net/http"
	"strings"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/middleware"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
)

// CreateTachoHandler crea un nuevo tacho en MySQL y Neo4j
// @Summary Crear un nuevo tacho
// @Description Crea un tacho guardándolo tanto en MySQL como en Neo4j
// @Tags Tachos
// @Accept json
// @Produce json
// @Param tacho body services.CreateTachoRequest true "Datos del tacho a crear"
// @Success 201 {object} services.CreateTachoResponse "Tacho creado exitosamente"
// @Failure 400 {object} map[string]string "Datos de entrada inválidos"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /tachos [post]
func CreateTachoHandler(c *gin.Context) {
	var request services.CreateTachoRequest

	// Validar y bind del JSON
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	// Crear el tacho usando el servicio
	response, err := services.CreateTacho(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Actualizar métricas después de crear exitosamente
	// Incrementar contador de tachos para la zona (asumiendo que el barrio es la zona)
	middleware.UpdateTachoCapacidad(string(rune(response.TachoID)), request.Barrio, request.Capacidad)
	middleware.UpdateTachoPrioridad(string(rune(response.TachoID)), request.Barrio, float64(request.Prioridad))

	c.JSON(http.StatusCreated, response)
}

// DeleteTachoHandler elimina un tacho de MySQL y Neo4j
// @Summary Eliminar un tacho
// @Description Elimina un tacho tanto de MySQL como de Neo4j usando query parameters (custom_id O direccion+barrio)
// @Tags Tachos
// @Accept json
// @Produce json
// @Param custom_id query string false "ID personalizado del tacho (direccion|barrio)"
// @Param direccion query string false "Dirección del tacho (requiere también barrio)"
// @Param barrio query string false "Barrio del tacho (requerido si se pasa direccion)"
// @Success 200 {object} map[string]string "Tacho eliminado exitosamente"
// @Failure 400 {object} map[string]string "Parámetros inválidos"
// @Failure 404 {object} map[string]string "Tacho no encontrado en ninguna base"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /tachos [delete]
func DeleteTachoHandler(c *gin.Context) {
	// Obtener parámetros de query
	customID := c.Query("custom_id")
	direccion := c.Query("direccion")
	barrio := c.Query("barrio")

	var finalCustomID string

	// Validar que se proporcione al menos una opción válida
	if customID != "" {
		// Opción 1: usar custom_id directamente
		finalCustomID = customID
	} else if direccion != "" && barrio != "" {
		// Opción 2: construir custom_id desde direccion + barrio
		finalCustomID = direccion + "|" + barrio
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Debe proporcionar 'custom_id' O ('direccion' + 'barrio')",
			"examples": map[string]string{
				"opcion1": "?custom_id=Av Corrientes 1234|CHACARITA",
				"opcion2": "?direccion=Av Corrientes 1234&barrio=CHACARITA",
			},
		})
		return
	}

	// Eliminar el tacho usando el servicio simplificado
	err := services.DeleteTacho(finalCustomID)
	if err != nil {
		if strings.Contains(err.Error(), "no se pudo eliminar de ninguna base de datos") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tacho no encontrado en ninguna base de datos"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Tacho eliminado exitosamente",
		"custom_id": finalCustomID,
	})
}

// GetAllTachosHandler obtiene todos los tachos con información completa
// @Summary Obtener todos los tachos
// @Description Devuelve todos los tachos con barrio, dirección, latitud, longitud, estado y capacidad
// @Tags Tachos
// @Accept json
// @Produce json
// @Success 200 {array} services.TachoCompleto "Lista de todos los tachos"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /tachos [get]
func GetAllTachosHandler(c *gin.Context) {
	// Obtener todos los tachos usando el servicio
	tachos, err := services.GetAllTachos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tachos": tachos,
		"total":  len(tachos),
	})
}
