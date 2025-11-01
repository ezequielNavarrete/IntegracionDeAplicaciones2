package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
)

var ctx = context.Background()

// GeneratePersonasFromRoutes crea personas en Redis por cada ruta calculada
func GeneratePersonasFromRoutes() error {
	log.Println("🧑‍🤝‍🧑 Generando personas a partir de las rutas cacheadas...")

	// Primero limpiar personas existentes
	err := clearExistingPersonas()
	if err != nil {
		return fmt.Errorf("error clearing existing personas: %w", err)
	}

	// Limpiar mapeos email->userID antiguos (excepto los dummy originales)
	err = clearOldEmailMappings()
	if err != nil {
		log.Printf("⚠️  Error limpiando emails antiguos: %v", err)
	}

	// Obtener todas las keys de rutas de Redis
	keys, err := config.RedisClient.Keys(ctx, "barrio_*_ruta_*").Result()
	if err != nil {
		return fmt.Errorf("error getting route keys: %w", err)
	}

	if len(keys) == 0 {
		log.Println("⚠️  No se encontraron rutas en cache, no se pueden generar personas")
		return nil
	}

	log.Printf("📊 Se encontraron %d rutas en cache\n", len(keys))
	log.Printf("📧 Se generarán emails automáticos para todas las rutas\n")

	personaCounter := 1

	// Iterar sobre cada ruta y crear una persona
	for _, key := range keys {
		// Extraer route number del key (formato: barrio_X_ruta_Y)
		var neighborhoodID, routeNumber int
		fmt.Sscanf(key, "barrio_%d_ruta_%d", &neighborhoodID, &routeNumber)

		// Obtener la ruta desde Redis
		routeJSON, err := config.RedisClient.Get(ctx, key).Result()
		if err != nil {
			log.Printf("⚠️  Error obteniendo ruta %s: %v", key, err)
			continue
		}

		// Parsear la ruta para obtener neighborhood, route_number y truck_id
		var route SimplifiedRoute
		if err := json.Unmarshal([]byte(routeJSON), &route); err != nil {
			log.Printf("⚠️  Error parseando ruta %s: %v", key, err)
			continue
		}

		// Generar nombre de persona y email
		personaNombre := fmt.Sprintf("Conductor_B%d_R%d", neighborhoodID, routeNumber)
		personaEmail := fmt.Sprintf("conductor.b%d.r%d@empresa.com", neighborhoodID, routeNumber)

		// Crear persona
		personKey := fmt.Sprintf("persona:%d", personaCounter)
		personData := map[string]interface{}{
			"id":              fmt.Sprintf("%d", personaCounter),
			"nombre":          personaNombre,
			"email":           personaEmail,
			"neighborhood_id": fmt.Sprintf("%d", neighborhoodID),
			"route_number":    fmt.Sprintf("%d", routeNumber),
			"truck_id":        fmt.Sprintf("%d", route.TruckID),
			"created_at":      time.Now().Format(time.RFC3339),
		}

		// Guardar hash en Redis usando HSet (compatible con todas las versiones)
		err = config.RedisClient.HSet(ctx, personKey, personData).Err()
		if err != nil {
			log.Printf("❌ Error creando persona %s: %v", personKey, err)
			continue
		}

		// Agregar a la lista de personas
		err = config.RedisClient.LPush(ctx, "personas", personKey).Err()
		if err != nil {
			log.Printf("❌ Error agregando %s a lista: %v", personKey, err)
			continue
		}

		// Crear mapeo email -> userID en Redis
		err = config.RedisClient.Set(ctx, personaEmail, fmt.Sprintf("%d", personaCounter), 0).Err()
		if err != nil {
			log.Printf("⚠️  Error creando mapeo email->userID para %s: %v", personaEmail, err)
		}

		log.Printf("✅ Creada %s - %s | Email: %s | Barrio:%d, Ruta:%d, Camión:%d",
			personKey, personaNombre, personaEmail, neighborhoodID, routeNumber, route.TruckID)

		personaCounter++
	}

	totalPersonas := personaCounter - 1
	log.Printf("✅ Generación completada: %d personas creadas con sus emails\n", totalPersonas)
	log.Printf("📧 Emails generados con formato: conductor.bX.rY@empresa.com\n")

	return nil
}

// clearExistingPersonas limpia las personas existentes en Redis
func clearExistingPersonas() error {
	log.Println("🧹 Limpiando personas existentes...")

	// Obtener la lista de personas
	personas, err := config.RedisClient.LRange(ctx, "personas", 0, -1).Result()
	if err != nil {
		return err
	}

	// Eliminar cada persona
	for _, personKey := range personas {
		if err := config.RedisClient.Del(ctx, personKey).Err(); err != nil {
			log.Printf("⚠️  Error eliminando %s: %v", personKey, err)
		}
	}

	// Eliminar la lista de personas
	if err := config.RedisClient.Del(ctx, "personas").Err(); err != nil {
		return err
	}

	log.Printf("✅ Se limpiaron %d personas\n", len(personas))
	return nil
}

// clearOldEmailMappings limpia los mapeos email->userID de personas generadas anteriormente
// Mantiene los emails dummy originales (eze@example.com, etc.)
func clearOldEmailMappings() error {
	log.Println("🧹 Limpiando mapeos de email antiguos...")

	// Buscar todas las keys que parecen emails de conductores generados
	emailKeys, err := config.RedisClient.Keys(ctx, "conductor.*@empresa.com").Result()
	if err != nil {
		return err
	}

	for _, emailKey := range emailKeys {
		if err := config.RedisClient.Del(ctx, emailKey).Err(); err != nil {
			log.Printf("⚠️  Error eliminando mapeo %s: %v", emailKey, err)
		}
	}

	log.Printf("✅ Se limpiaron %d mapeos de email\n", len(emailKeys))
	return nil
}
