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
