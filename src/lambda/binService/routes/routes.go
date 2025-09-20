package routes

import (
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/handlers"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine) {
	// Swagger documentation endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API endpoints
	r.GET("/ruta-optima/:zonaID", handlers.GetRutaHandler)
	r.POST("/enviar-emergencia", handlers.SendEmergencyHandler)
}
