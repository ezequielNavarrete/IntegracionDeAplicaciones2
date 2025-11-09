package handlers

import (
	"net/http"
	"strconv"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
)

// GetAllReclamosHandler obtiene todos los reclamos
// @Summary Obtener todos los reclamos
// @Description Devuelve una lista de todos los reclamos
// @Tags Reclamos
// @Accept json
// @Produce json
// @Success 200 {array} services.Reclamo "Lista de reclamos"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /reclamos [get]
func GetAllReclamosHandler(c *gin.Context) {
	reclamos, err := services.GetAllReclamos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reclamos": reclamos,
		"total":    len(reclamos),
	})
}

// GetReclamoByIDHandler obtiene un reclamo específico por ID
// @Summary Obtener un reclamo por ID
// @Description Devuelve un reclamo específico
// @Tags Reclamos
// @Accept json
// @Produce json
// @Param id path int true "ID del reclamo"
// @Success 200 {object} services.Reclamo "Reclamo encontrado"
// @Failure 400 {object} map[string]string "ID inválido"
// @Failure 404 {object} map[string]string "Reclamo no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /reclamos/{id} [get]
func GetReclamoByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	reclamoID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	reclamo, err := services.GetReclamoByID(reclamoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reclamo no encontrado"})
		return
	}

	c.JSON(http.StatusOK, reclamo)
}

// CreateReclamoHandler crea un nuevo reclamo
// @Summary Crear un nuevo reclamo
// @Description Crea un reclamo con la información proporcionada
// @Tags Reclamos
// @Accept json
// @Produce json
// @Param reclamo body services.CreateReclamoRequest true "Datos del reclamo"
// @Success 201 {object} services.CreateReclamoResponse "Reclamo creado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /reclamos [post]
func CreateReclamoHandler(c *gin.Context) {
	var request services.CreateReclamoRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	response, err := services.CreateReclamo(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// DeleteReclamoHandler elimina un reclamo
// @Summary Eliminar un reclamo
// @Description Elimina un reclamo por su ID
// @Tags Reclamos
// @Accept json
// @Produce json
// @Param id path int true "ID del reclamo"
// @Success 200 {object} map[string]string "Reclamo eliminado exitosamente"
// @Failure 400 {object} map[string]string "ID inválido"
// @Failure 404 {object} map[string]string "Reclamo no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /reclamos/{id} [delete]
func DeleteReclamoHandler(c *gin.Context) {
	idStr := c.Param("id")
	reclamoID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := services.DeleteReclamo(reclamoID); err != nil {
		if err.Error() == "reclamo not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reclamo no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reclamo eliminado exitosamente"})
}

// UpdateReclamoEstadoHandler actualiza el estado de un reclamo
// @Summary Actualizar estado de un reclamo
// @Description Actualiza el estado de un reclamo específico
// @Tags Reclamos
// @Accept json
// @Produce json
// @Param id path int true "ID del reclamo"
// @Param estado body map[string]string true "Nuevo estado"
// @Success 200 {object} map[string]string "Estado actualizado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 404 {object} map[string]string "Reclamo no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /reclamos/{id}/estado [put]
func UpdateReclamoEstadoHandler(c *gin.Context) {
	idStr := c.Param("id")
	reclamoID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var body struct {
		Estado string `json:"estado" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Estado requerido"})
		return
	}

	if err := services.UpdateReclamoEstado(reclamoID, body.Estado); err != nil {
		if err.Error() == "reclamo not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reclamo no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Estado actualizado exitosamente",
		"estado":  body.Estado,
	})
}
