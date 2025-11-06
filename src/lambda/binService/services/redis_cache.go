package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
)

// TTL por defecto en segundos (puede ser override por env REDIS_TTL_SECONDS)
var defaultTTL = 300 * time.Second

func init() {
	if v := os.Getenv("REDIS_TTL_SECONDS"); v != "" {
		if secs, err := time.ParseDuration(v + "s"); err == nil {
			defaultTTL = secs
		}
	}
}

// GetCachedRoute intenta obtener la ruta cacheada para una zona
func GetCachedRoute(zonaID int) ([]Point, error) {
	if config.RedisClient == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	ctx := context.Background()
	key := fmt.Sprintf("ruta:zona:%d", zonaID)
	val, err := config.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var points []Point
	if err := json.Unmarshal([]byte(val), &points); err != nil {
		return nil, err
	}

	return points, nil
}

// SetCachedRoute guarda la ruta en Redis con TTL
func SetCachedRoute(zonaID int, points []Point) error {
	if config.RedisClient == nil {
		return fmt.Errorf("redis client not available")
	}

	ctx := context.Background()
	key := fmt.Sprintf("ruta:zona:%d", zonaID)
	b, err := json.Marshal(points)
	if err != nil {
		return err
	}

	return config.RedisClient.Set(ctx, key, b, defaultTTL).Err()
}

// SetSimplifiedRoute guarda una ruta simplificada en Redis con formato barrio_X_ruta_Y
func SetSimplifiedRoute(route SimplifiedRoute, routeNumber int) error {
	if config.RedisClient == nil {
		return fmt.Errorf("redis client not available")
	}

	ctx := context.Background()
	key := fmt.Sprintf("barrio_%d_ruta_%d", route.Neighborhood, routeNumber)

	// Convertir la ruta a JSON
	b, err := json.Marshal(route)
	if err != nil {
		return fmt.Errorf("error marshaling route: %w", err)
	}

	// Guardar en Redis con TTL
	if err := config.RedisClient.Set(ctx, key, b, defaultTTL).Err(); err != nil {
		return fmt.Errorf("error setting route in Redis: %w", err)
	}

	return nil
}

// GetSimplifiedRoute obtiene una ruta simplificada desde Redis
func GetSimplifiedRoute(neighborhood int, routeNumber int) (*SimplifiedRoute, error) {
	if config.RedisClient == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	ctx := context.Background()
	key := fmt.Sprintf("barrio_%d_ruta_%d", neighborhood, routeNumber)

	val, err := config.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var route SimplifiedRoute
	if err := json.Unmarshal([]byte(val), &route); err != nil {
		return nil, fmt.Errorf("error unmarshaling route: %w", err)
	}

	return &route, nil
}

// GetAllRoutesForNeighborhood obtiene todas las rutas de un barrio
func GetAllRoutesForNeighborhood(neighborhood int) ([]SimplifiedRoute, error) {
	if config.RedisClient == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	ctx := context.Background()
	pattern := fmt.Sprintf("barrio_%d_ruta_*", neighborhood)

	keys, err := config.RedisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("error getting keys: %w", err)
	}

	routes := make([]SimplifiedRoute, 0, len(keys))
	for _, key := range keys {
		val, err := config.RedisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var route SimplifiedRoute
		if err := json.Unmarshal([]byte(val), &route); err != nil {
			continue
		}

		routes = append(routes, route)
	}

	return routes, nil
}

// NeighborhoodRoutesInfo representa la información de rutas por neighborhood
type NeighborhoodRoutesInfo struct {
	Neighborhood int   `json:"neighborhood"`
	Routes       []int `json:"routes"`
}

// GetAllNeighborhoodsWithRoutes obtiene todos los neighborhoods y sus rutas disponibles
func GetAllNeighborhoodsWithRoutes() ([]NeighborhoodRoutesInfo, error) {
	if config.RedisClient == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	ctx := context.Background()
	pattern := "barrio_*_ruta_*"

	keys, err := config.RedisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("error getting keys: %w", err)
	}

	// Map para agrupar rutas por neighborhood
	neighborhoodMap := make(map[int]map[int]bool)

	// Parsear las keys para extraer neighborhood y route number
	for _, key := range keys {
		var neighborhood, routeNum int
		// Formato: barrio_X_ruta_Y
		if _, err := fmt.Sscanf(key, "barrio_%d_ruta_%d", &neighborhood, &routeNum); err == nil {
			if neighborhoodMap[neighborhood] == nil {
				neighborhoodMap[neighborhood] = make(map[int]bool)
			}
			neighborhoodMap[neighborhood][routeNum] = true
		}
	}

	// Convertir el map a slice
	result := make([]NeighborhoodRoutesInfo, 0, len(neighborhoodMap))
	for neighborhood, routesMap := range neighborhoodMap {
		routes := make([]int, 0, len(routesMap))
		for routeNum := range routesMap {
			routes = append(routes, routeNum)
		}
		
		result = append(result, NeighborhoodRoutesInfo{
			Neighborhood: neighborhood,
			Routes:       routes,
		})
	}

	return result, nil
}
