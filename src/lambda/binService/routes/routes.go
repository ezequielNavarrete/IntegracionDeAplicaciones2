package routes

import (
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/handlers"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRoutes(r *gin.Engine) {
	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API endpoints existentes
	r.GET("/ruta-optima", handlers.GetRutaHandlerByHeader)
	r.GET("/ruta-optima/:zonaID", handlers.GetRutaHandler)
	r.POST("/enviar-emergencia", handlers.SendEmergencyHandler)

	// Endpoints V2 de ruta óptima (basados en usuario autenticado)
	r.GET("/v2/ruta-optima", handlers.GetRutaHandlerV2)                   // Con JWT
	r.GET("/v2/ruta-optima-by-header", handlers.GetRutaHandlerV2ByHeader) // Con header X-User-Email

	// Endpoints V2 de navegación punto a punto
	r.GET("/v2/ruta-navegacion/:neighborhood/:route", handlers.GetRouteNavigationHandler)  // Navegación por barrio/ruta
	r.GET("/v2/ruta-navegacion/:neighborhood/:route/start", handlers.GetRouteStartHandler) // Iniciar navegación
	r.GET("/v2/ruta-navegacion-by-header", handlers.GetRouteNavigationByHeaderHandler)     // Navegación por email
	r.GET("/v2/ruta-navegacion-by-header/start", handlers.GetRouteStartByHeaderHandler)    // Iniciar navegación por email

	// Nuevos endpoints para personas (Redis)
	r.GET("/personas", handlers.GetAllPersonas)
	r.GET("/personas/:id", handlers.GetPersonaByID)
	r.GET("/personas/neighborhood/:neighborhood", handlers.GetPersonasByNeighborhood)
	r.GET("/personas/emails", handlers.GetPersonasWithEmails)          // Nuevo: lista con emails
	r.POST("/personas", handlers.CreatePersonaHandler)                 // Crear persona (Redis)
	r.DELETE("/personas/:id", handlers.DeletePersonaHandler)           // Eliminar persona (Redis)
	r.POST("/personas/regenerate", handlers.RegeneratePersonasHandler) // Nuevo: regenerar personas

	// Endpoints para tachos
	r.GET("/tachos", handlers.GetAllTachosHandler) // Obtener todos los tachos
	r.POST("/tachos", handlers.CreateTachoHandler)
	r.DELETE("/tachos", handlers.DeleteTachoHandler) // Cambiado para usar query parameters
	r.PUT("/tachos/:id_tacho/capacidad", handlers.UpdateCapacidadTachoHandler)
	r.PUT("/tachos/:id_tacho/prioridad", handlers.UpdatePrioridadTachoHandler)

	// Endpoints para camiones
	r.GET("/camiones", handlers.GetAllCamionesHandler)    // Obtener todos los camiones (MySQL)
	r.GET("/camiones/:id", handlers.GetCamionByIDHandler) // Obtener camión por ID (MySQL)
	r.POST("/camiones", handlers.CreateCamionHandler)     // Crear camión (Redis temporal)
	r.DELETE("/camiones/:id", handlers.DeleteCamionHandler) // Eliminar camión (Redis temporal)

	// Endpoints para centros
	r.GET("/centros", handlers.GetAllCentrosHandler)     // Obtener todos los centros con JOIN MySQL + Neo4j
	r.GET("/centros/:id", handlers.GetCentroByIDHandler) // Obtener centro por ID con JOIN MySQL + Neo4j

	// Endpoints para cron jobs
	r.POST("/cron/update-routes", handlers.UpdateAllRoutesHandler) // Actualizar todas las rutas en caché

	// Endpoints para consultar rutas cacheadas
	r.GET("/routes/neighborhoods", handlers.GetAllNeighborhoodsWithRoutesHandler)                    // Listar todos los neighborhoods con sus rutas
	r.GET("/routes/neighborhood/:neighborhood", handlers.GetRoutesByNeighborhoodHandler)             // Obtener todas las rutas de un barrio
	r.GET("/routes/neighborhood/:neighborhood/route/:routeNumber", handlers.GetSpecificRouteHandler) // Obtener una ruta específica
}
