package routes

import (
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// API endpoints existentes
	r.GET("/ruta-optima", handlers.GetRutaHandlerByHeader)
	r.GET("/ruta-optima/:zonaID", handlers.GetRutaHandler)
	r.POST("/enviar-emergencia", handlers.SendEmergencyHandler)

	// Nuevos endpoints para personas (Redis)
	r.GET("/personas", handlers.GetAllPersonas)
	r.GET("/personas/:id", handlers.GetPersonaByID)
	r.GET("/personas/zona/:zona", handlers.GetPersonasByZona)

	// Endpoints para tachos
	r.GET("/tachos", handlers.GetAllTachosHandler)    // Obtener todos los tachos
	r.POST("/tachos", handlers.CreateTachoHandler)
	r.DELETE("/tachos", handlers.DeleteTachoHandler) // Cambiado para usar query parameters
	r.PUT("/tachos/:id_tacho/capacidad", handlers.UpdateCapacidadTachoHandler)
	r.PUT("/tachos/:id_tacho/prioridad", handlers.UpdatePrioridadTachoHandler)
}
