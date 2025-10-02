package main

// @title IntegracionDeAplicaciones2 API
// @version 1.0
// @description API para gestión de rutas y emergencias
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	_ "github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/docs" // Import generated docs
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/middleware"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/routes"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Connect to MySQL
	config.ConnectDatabase()

	// Connect to Redis and initialize data
	config.ConnectRedis()

	// Initialize Neo4j driver pool
	_, err := config.GetNeo4jDriver()
	if err != nil {
		log.Printf("Warning: Neo4j connection failed: %v", err)
	}

	// Close Neo4j driver on app shutdown
	defer config.CloseNeo4jDriver()

	// start Gin server
	r := gin.Default()

	// Add Prometheus metrics middleware
	r.Use(middleware.PrometheusMiddleware())

	// Configuración de CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // tu frontend
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Setup routes (API endpoints only)
	routes.SetupRoutes(r)

	// Configure Swagger dynamically
	swaggerHost := os.Getenv("SWAGGER_HOST")
	if swaggerHost == "" {
		swaggerHost = "localhost:8080"
	}

	swaggerScheme := os.Getenv("SWAGGER_SCHEME")
	if swaggerScheme == "" {
		swaggerScheme = "http"
	}

	// Swagger endpoint with dynamic config
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL(swaggerScheme+"://"+swaggerHost+"/swagger/doc.json"),
	))

	// Get port from environment
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
