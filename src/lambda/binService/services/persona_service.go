package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
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

// =====================
// Creación/Eliminación individual de personas en Redis
// =====================

// Persona representa la entidad en la capa de servicios
type Persona struct {
	ID             string
	Nombre         string
	Email          string
	NeighborhoodID int
	RouteNumber    int
	TruckID        int
}

// CreatePersonaRequest request esperado para crear persona
type CreatePersonaRequest struct {
	Nombre         string `json:"nombre" binding:"required,min=2"`
	Email          string `json:"email" binding:"required,email"`
	NeighborhoodID int    `json:"neighborhood_id" binding:"required,min=1"`
	RouteNumber    int    `json:"route_number" binding:"required,min=0"`
	TruckID        int    `json:"truck_id" binding:"required,min=0"`
}

// CreatePersonaRedis crea una persona con el esquema actual:
// - Hash: persona:<id>
// - Lista: personas (contiene el valor personKey, p.ej. "persona:3")
// - Mapeo email -> id (clave email sin prefijo, siguiendo patrón existente)
func CreatePersonaRedis(req CreatePersonaRequest) (*Persona, error) {
	if config.RedisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	// Determinar próximo ID buscando el máximo en la lista actual
	keys, err := config.RedisClient.LRange(ctx, "personas", 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("error leyendo lista personas: %w", err)
	}
	nextID := 1
	if len(keys) > 0 {
		ids := make([]int, 0, len(keys))
		for _, k := range keys {
			// k formato: "persona:N"
			parts := strings.Split(k, ":")
			if len(parts) == 2 {
				if n, convErr := strconv.Atoi(parts[1]); convErr == nil {
					ids = append(ids, n)
				}
			}
		}
		if len(ids) > 0 {
			sort.Ints(ids)
			nextID = ids[len(ids)-1] + 1
		}
	}

	personKey := fmt.Sprintf("persona:%d", nextID)
	now := time.Now().Format(time.RFC3339)

	data := map[string]interface{}{
		"id":              strconv.Itoa(nextID),
		"nombre":          req.Nombre,
		"email":           req.Email,
		"neighborhood_id": strconv.Itoa(req.NeighborhoodID),
		"route_number":    strconv.Itoa(req.RouteNumber),
		"truck_id":        strconv.Itoa(req.TruckID),
		"created_at":      now,
	}

	if err := config.RedisClient.HSet(ctx, personKey, data).Err(); err != nil {
		return nil, fmt.Errorf("error guardando persona: %w", err)
	}
	if err := config.RedisClient.LPush(ctx, "personas", personKey).Err(); err != nil {
		return nil, fmt.Errorf("error agregando a lista personas: %w", err)
	}
	if err := config.RedisClient.Set(ctx, req.Email, strconv.Itoa(nextID), 0).Err(); err != nil {
		log.Printf("⚠️  Error creando mapeo email->id para %s: %v", req.Email, err)
	}

	return &Persona{
		ID:             strconv.Itoa(nextID),
		Nombre:         req.Nombre,
		Email:          req.Email,
		NeighborhoodID: req.NeighborhoodID,
		RouteNumber:    req.RouteNumber,
		TruckID:        req.TruckID,
	}, nil
}

// DeletePersonaRedis elimina persona:<id> y la referencia en la lista y el mapeo email->id
func DeletePersonaRedis(id int) error {
	if config.RedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	if id <= 0 {
		return fmt.Errorf("id inválido")
	}

	personKey := fmt.Sprintf("persona:%d", id)
	exists, err := config.RedisClient.Exists(ctx, personKey).Result()
	if err != nil {
		return fmt.Errorf("error verificando existencia: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("persona %d not found", id)
	}

	// Obtener email para limpiar mapeo
	email, _ := config.RedisClient.HGet(ctx, personKey, "email").Result()

	if err := config.RedisClient.Del(ctx, personKey).Err(); err != nil {
		return fmt.Errorf("error eliminando persona: %w", err)
	}
	if err := config.RedisClient.LRem(ctx, "personas", 0, personKey).Err(); err != nil {
		return fmt.Errorf("error removiendo de lista: %w", err)
	}
	if email != "" {
		if err := config.RedisClient.Del(ctx, email).Err(); err != nil {
			log.Printf("⚠️  Error eliminando mapeo email %s: %v", email, err)
		}
	}

	return nil
}
