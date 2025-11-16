package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/middleware"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
)

// UpdateAllRoutesHandler actualiza todas las rutas en caché
// @Summary Actualizar todas las rutas en caché
// @Description Calcula y almacena en caché las rutas óptimas para todas las zonas
// @Tags Cron
// @Accept json
// @Produce json
// @Param Authorization header string true "Clave de API para cron jobs"
// @Success 200 {object} map[string]interface{} "Resultados de la actualización"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /cron/update-routes [post]
func UpdateAllRoutesHandler(c *gin.Context) {
	separator := strings.Repeat("=", 80)
	fmt.Println("\n" + separator)
	fmt.Println("🔄 [CRON] Iniciando actualización de rutas...")
	fmt.Println(separator)

	// Verificar autorización
	apiKey := c.GetHeader("Authorization")
	expectedKey := os.Getenv("CRON_API_KEY")

	fmt.Printf("[CRON] 🔑 Verificando autorización...\n")
	fmt.Printf("[CRON]    Header recibido: %s\n", apiKey[:min(20, len(apiKey))]+"...")
	fmt.Printf("[CRON]    Expected key: %s\n", expectedKey[:min(20, len(expectedKey))]+"...")

	if expectedKey == "" || apiKey != expectedKey {
		fmt.Printf("[CRON] ❌ Autorización FALLIDA\n")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
		return
	}
	fmt.Printf("[CRON] ✅ Autorización exitosa\n\n")

	start := time.Now()

	// 1. Obtener camiones disponibles (filtrar por estado "Operativo" o similar)
	fmt.Printf("[CRON] 📊 Paso 1: Obteniendo camiones disponibles...\n")
	camionesResp, err := services.GetAllCamiones()
	if err != nil {
		fmt.Printf("[CRON] ❌ Error obteniendo camiones: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error obteniendo camiones: %v", err)})
		return
	}
	fmt.Printf("[CRON] ✅ Total de camiones en DB: %d\n", len(camionesResp.Camiones))

	// Convertir camiones a formato para rutas
	availableTrucks := make([]services.TruckForRouteRequest, 0)
	for _, camion := range camionesResp.Camiones {
		// Filtrar solo camiones operativos (ajustar según tu lógica)
		if camion.TipoEstado == "Operativo" || camion.TipoEstado == "Disponible" {
			// Asignar capacidad según tipo (ajustar según tu lógica)
			capacity := 8000 // Default
			if camion.NombreTipo == "Reciclaje" {
				capacity = 6000
			} else if camion.NombreTipo == "Especial" {
				capacity = 7000
			}

			availableTrucks = append(availableTrucks, services.TruckForRouteRequest{
				ID:       camion.IDCamion,
				Capacity: capacity,
				Type:     camion.NombreTipo,
			})
			fmt.Printf("[CRON]    ✓ Camión %d: %s, %s, %dkg\n", camion.IDCamion, camion.NombreTipo, camion.TipoEstado, capacity)
		}
	}

	if len(availableTrucks) == 0 {
		fmt.Printf("[CRON] ❌ No hay camiones disponibles\n")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No hay camiones disponibles"})
		return
	}
	fmt.Printf("[CRON] ✅ Camiones operativos: %d\n\n", len(availableTrucks))

	// 2. Obtener estadísticas de tachos por barrio desde MongoDB
	fmt.Printf("[CRON] 📊 Paso 2: Consultando MongoDB para estadísticas de barrios...\n")
	neighborhoodStats, err := services.GetBinsStatsByNeighborhood()
	if err != nil {
		fmt.Printf("[CRON] ❌ Error obteniendo estadísticas: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error obteniendo estadísticas de barrios: %v", err)})
		return
	}

	if len(neighborhoodStats) == 0 {
		fmt.Printf("[CRON] ❌ No se encontraron barrios con tachos\n")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se encontraron barrios con tachos"})
		return
	}

	fmt.Printf("[CRON] ✅ Barrios encontrados: %d\n", len(neighborhoodStats))
	for _, stat := range neighborhoodStats {
		fmt.Printf("[CRON]    Barrio %d: %d tachos\n", stat.Neighborhood, stat.TotalBins)
	}
	fmt.Println()

	// 3. Distribuir camiones entre barrios
	fmt.Printf("[CRON] 📊 Paso 3: Distribuyendo camiones entre barrios...\n")
	assignments, err := services.DistributeTrucksAcrossNeighborhoods(availableTrucks, neighborhoodStats)
	if err != nil {
		fmt.Printf("[CRON] ❌ Error distribuyendo: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error distribuyendo camiones: %v", err)})
		return
	}

	fmt.Printf("[CRON] ✅ Distribución completada:\n")
	for _, assignment := range assignments {
		fmt.Printf("[CRON]    Barrio %d: %d camiones asignados\n", assignment.Neighborhood, len(assignment.Trucks))
	}
	fmt.Println()

	// 4. Calcular rutas para cada barrio y guardar en Redis
	fmt.Printf("[CRON] 📊 Paso 4: Calculando rutas y guardando en Redis...\n")
	totalRoutesProcessed := 0
	successCount := 0
	errorCount := 0
	errorMessages := make([]string, 0)

	for _, assignment := range assignments {
		fmt.Printf("[CRON] 🚀 Procesando barrio %d con %d camiones...\n", assignment.Neighborhood, len(assignment.Trucks))

		// Calcular rutas para este barrio
		routes, err := services.CalculateRoutesForNeighborhood(assignment.Neighborhood, assignment.Trucks)
		if err != nil {
			fmt.Printf("[CRON] ❌ Error calculando rutas para barrio %d: %v\n", assignment.Neighborhood, err)
			errorCount++
			errorMessages = append(errorMessages, fmt.Sprintf("Barrio %d: %v", assignment.Neighborhood, err))
			continue
		}

		fmt.Printf("[CRON] ✅ Rutas calculadas para barrio %d: %d rutas\n", assignment.Neighborhood, len(routes))

		// Guardar cada ruta en Redis
		for i, route := range routes {
			routeNumber := i + 1
			fmt.Printf("[CRON]    💾 Guardando ruta %d (barrio_%d_ruta_%d)...\n", routeNumber, assignment.Neighborhood, routeNumber)

			if err := services.SetSimplifiedRoute(route, routeNumber); err != nil {
				fmt.Printf("[CRON]    ❌ Error guardando ruta: %v\n", err)
				errorCount++
				errorMessages = append(errorMessages, fmt.Sprintf("Error guardando barrio %d ruta %d: %v", assignment.Neighborhood, routeNumber, err))
			} else {
				fmt.Printf("[CRON]    ✅ Ruta guardada exitosamente\n")
				successCount++
				totalRoutesProcessed++
			}
		}
		fmt.Println()
	}

	duration := time.Since(start)
	middleware.ObserveCronUpdateTime(duration.Seconds())

	// 5. Actualizar el timestamp de última recolección
	fmt.Printf("[CRON] 📊 Paso 5: Actualizando horario de última recolección...\n")
	if err := services.UpdateLastCollectionTime(); err != nil {
		fmt.Printf("[CRON] ⚠️  Error actualizando horario: %v\n", err)
		// No falla el proceso completo si falla la actualización del horario
	} else {
		fmt.Printf("[CRON] ✅ Horario actualizado exitosamente\n")
	}
	fmt.Println()

	// 6. Generar personas a partir de las rutas
	fmt.Printf("[CRON] 📊 Paso 6: Generando personas a partir de las rutas...\n")
	if err := services.GeneratePersonasFromRoutes(); err != nil {
		fmt.Printf("[CRON] ⚠️  Error generando personas: %v\n", err)
		// No falla el proceso completo si falla la generación de personas
	} else {
		fmt.Printf("[CRON] ✅ Personas generadas exitosamente\n")
	}
	fmt.Println()

	fmt.Println(separator)
	fmt.Printf("✅ [CRON] Actualización completada en %.2f segundos\n", duration.Seconds())
	fmt.Printf("   📊 Total de rutas procesadas: %d\n", totalRoutesProcessed)
	fmt.Printf("   ✅ Exitosas: %d\n", successCount)
	fmt.Printf("   ❌ Errores: %d\n", errorCount)
	fmt.Println(separator + "\n")

	c.JSON(http.StatusOK, gin.H{
		"total_neighborhoods":  len(assignments),
		"total_trucks_used":    len(availableTrucks),
		"total_routes_created": totalRoutesProcessed,
		"success_count":        successCount,
		"error_count":          errorCount,
		"errors":               errorMessages,
		"duration_seconds":     duration.Seconds(),
		"assignments":          assignments,
	})
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
