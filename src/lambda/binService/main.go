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

// @host localhost:8080
// @BasePath /
// @schemes http

import (
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	_ "github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/docs" // Import generated docs
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Connect to MySQL
	config.ConnectDatabase()

	// start Gin server
	r := gin.Default()

	// record the routes
	routes.SetupRoutes(r)

	// Run server on port 8080
	r.Run(":8080")
}
