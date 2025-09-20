package routes

import (
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/demo/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.GET("/ruta-optima/:zonaID", handlers.GetRutaHandler)

	r.POST("/enviar-emergencia", handlers.SendEmergencyHandler)
}
