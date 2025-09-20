package main

import (
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/demo/config"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/demo/routes"
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
