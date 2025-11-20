package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
)

// TruckAssignment representa la asignación de camiones a un barrio
type TruckAssignment struct {
	Neighborhood int                    `json:"neighborhood"`
	Trucks       []TruckForRouteRequest `json:"trucks"`
	TotalBins    int                    `json:"total_bins"`
}

// TruckForRouteRequest representa un camión para la solicitud de ruta
type TruckForRouteRequest struct {
	ID       int    `json:"id"`
	Capacity int    `json:"capacity"`
	Type     string `json:"type"`
}

// RouteRequest representa la solicitud completa para calcular rutas
type RouteRequest struct {
	Neighborhood            int                    `json:"neighborhood"`
	Trucks                  []TruckForRouteRequest `json:"trucks"`
	MinPriority             int                    `json:"min_priority"`
	MaxBinsPerTruck         int                    `json:"max_bins_per_truck"`
	RespectStreetDirection  bool                   `json:"respect_street_direction"`
	AllowedStatuses         []string               `json:"allowed_statuses"`
	UrgentMultiplier        float64                `json:"urgent_multiplier"`
	IncludeBorderConnectors bool                   `json:"include_border_connectors"`
	SequenceStrategy        string                 `json:"sequence_strategy"`
	PriorityBias            float64                `json:"priority_bias"`
}

// SimplifiedRoute representa una ruta simplificada para guardar en Redis
type SimplifiedRoute struct {
	RouteID      string       `json:"route_id"`
	TruckID      int          `json:"truck_id"`
	Neighborhood int          `json:"neighborhood"`
	BinsCoords   []Coordinate `json:"bins_coords"` // Coordenadas de tachos en orden
	PathCoords   []Coordinate `json:"path_coords"` // Coordenadas del path completo
	TotalBins    int          `json:"total_bins"`
	TotalDistKm  float64      `json:"total_distance_km"`
}

// Coordinate representa una coordenada geográfica
type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// DistributeTrucksAcrossNeighborhoods distribuye camiones según cantidad de tachos
func DistributeTrucksAcrossNeighborhoods(availableTrucks []TruckForRouteRequest, neighborhoodStats []NeighborhoodStats) ([]TruckAssignment, error) {
	if len(availableTrucks) == 0 {
		return nil, fmt.Errorf("no available trucks")
	}

	if len(neighborhoodStats) == 0 {
		return nil, fmt.Errorf("no neighborhoods with bins")
	}

	// Ordenar barrios por cantidad de tachos (descendente)
	sort.Slice(neighborhoodStats, func(i, j int) bool {
		return neighborhoodStats[i].TotalBins > neighborhoodStats[j].TotalBins
	})

	// Calcular total de tachos
	totalBins := 0
	for _, stat := range neighborhoodStats {
		totalBins += stat.TotalBins
	}

	assignments := make([]TruckAssignment, 0, len(neighborhoodStats))

	// Primero asignar 1 camión a cada barrio
	trucksUsed := 0
	for i, stat := range neighborhoodStats {
		if trucksUsed >= len(availableTrucks) {
			break
		}

		assignments = append(assignments, TruckAssignment{
			Neighborhood: stat.Neighborhood,
			Trucks:       []TruckForRouteRequest{availableTrucks[trucksUsed]},
			TotalBins:    stat.TotalBins,
		})
		trucksUsed++

		// Si ya no hay más barrios, continuar con el siguiente
		if i >= len(neighborhoodStats)-1 {
			break
		}
	}

	// Distribuir camiones restantes proporcionalmente
	remainingTrucks := len(availableTrucks) - trucksUsed
	if remainingTrucks > 0 {
		for i := 0; i < remainingTrucks; i++ {
			// Encontrar el barrio con más tachos por camión asignado
			maxRatio := 0.0
			maxIdx := 0

			for j, assignment := range assignments {
				ratio := float64(assignment.TotalBins) / float64(len(assignment.Trucks))
				if ratio > maxRatio {
					maxRatio = ratio
					maxIdx = j
				}
			}

			// Asignar el siguiente camión al barrio con mayor ratio
			assignments[maxIdx].Trucks = append(assignments[maxIdx].Trucks, availableTrucks[trucksUsed])
			trucksUsed++
		}
	}

	return assignments, nil
}

// CalculateRoutesForNeighborhood calcula las rutas para un barrio específico
func CalculateRoutesForNeighborhood(neighborhood int, trucks []TruckForRouteRequest) ([]SimplifiedRoute, error) {
	// URL del servicio externo
	routeServiceURL := os.Getenv("ROUTE_SERVICE_URL")
	if routeServiceURL == "" {
		return nil, fmt.Errorf("ROUTE_SERVICE_URL not configured")
	}

	// Crear request
	request := RouteRequest{
		Neighborhood:            neighborhood,
		Trucks:                  trucks,
		MinPriority:             1,
		MaxBinsPerTruck:         50,
		RespectStreetDirection:  true,
		AllowedStatuses:         []string{"Bueno"},
		UrgentMultiplier:        1.0,
		IncludeBorderConnectors: true,
		SequenceStrategy:        "hybrid",
		PriorityBias:            0.5,
	}

	// Convertir a JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Hacer la solicitud HTTP
	resp, err := http.Post(routeServiceURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error calling route service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("route service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Leer respuesta
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Decodificar respuesta
	var routeResponse map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &routeResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Extraer rutas simplificadas
	routes, err := extractSimplifiedRoutes(routeResponse, neighborhood)
	if err != nil {
		return nil, fmt.Errorf("error extracting routes: %w", err)
	}

	return routes, nil
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractSimplifiedRoutes extrae las rutas simplificadas de la respuesta
func extractSimplifiedRoutes(response map[string]interface{}, neighborhood int) ([]SimplifiedRoute, error) {
	routesData, ok := response["routes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid routes format in response")
	}

	simplifiedRoutes := make([]SimplifiedRoute, 0, len(routesData))

	for _, routeData := range routesData {
		route, ok := routeData.(map[string]interface{})
		if !ok {
			continue
		}

		routeID, _ := route["route_id"].(string)

		// Extraer truck ID
		truckID := 0
		if truck, ok := route["truck"].(map[string]interface{}); ok {
			if id, ok := truck["id"].(float64); ok {
				truckID = int(id)
			}
		}

		// Extraer métricas
		totalDistKm := 0.0
		totalBins := 0
		if metrics, ok := route["metrics"].(map[string]interface{}); ok {
			if dist, ok := metrics["total_distance_km"].(float64); ok {
				totalDistKm = dist
			}
			if bins, ok := metrics["bins_count"].(float64); ok {
				totalBins = int(bins)
			}
		}

		// Extraer coordenadas del path primero para obtener depot
		pathCoords := make([]Coordinate, 0)
		if path, ok := route["path"].(map[string]interface{}); ok {
			if coordinates, ok := path["coordinates"].([]interface{}); ok {
				for _, coord := range coordinates {
					coordMap, ok := coord.(map[string]interface{})
					if !ok {
						continue
					}

					lat, _ := coordMap["lat"].(float64)
					lon, _ := coordMap["lon"].(float64)
					// Servicio externo devuelve coordenadas invertidas, las intercambiamos
					latReal, lonReal := SwapCoordinates(lat, lon)
					pathCoords = append(pathCoords, Coordinate{Lat: latReal, Lon: lonReal})
				}
			}
		}

		// Extraer coordenadas del depot (primera y última del path)
		var depotStart, depotEnd Coordinate
		if len(pathCoords) > 0 {
			depotStart = pathCoords[0]
			depotEnd = pathCoords[len(pathCoords)-1]
		}

		// Extraer coordenadas de tachos
		binsCoords := make([]Coordinate, 0)

		// Agregar depot al inicio
		if len(pathCoords) > 0 {
			binsCoords = append(binsCoords, depotStart)
		}

		// Agregar coordenadas de los tachos
		if bins, ok := route["bins"].([]interface{}); ok {
			for _, bin := range bins {
				binMap, ok := bin.(map[string]interface{})
				if !ok {
					continue
				}

				if coords, ok := binMap["coordinates"].(map[string]interface{}); ok {
					lat, _ := coords["lat"].(float64)
					lon, _ := coords["lon"].(float64)
					// Servicio externo devuelve coordenadas invertidas, las intercambiamos
					latReal, lonReal := SwapCoordinates(lat, lon)
					binsCoords = append(binsCoords, Coordinate{Lat: latReal, Lon: lonReal})
				}
			}
		}

		// Agregar depot al final
		if len(pathCoords) > 0 {
			binsCoords = append(binsCoords, depotEnd)
		}

		simplifiedRoutes = append(simplifiedRoutes, SimplifiedRoute{
			RouteID:      routeID,
			TruckID:      truckID,
			Neighborhood: neighborhood,
			BinsCoords:   binsCoords,
			PathCoords:   pathCoords,
			TotalBins:    totalBins,
			TotalDistKm:  totalDistKm,
		})
	}

	return simplifiedRoutes, nil
}
