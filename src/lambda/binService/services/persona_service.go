package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
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

	// Reiniciar secuencia de IDs para evitar colisiones posteriores
	if err := config.RedisClient.Set(ctx, "persona:id_seq", 0, 0).Err(); err != nil {
		return fmt.Errorf("error reiniciando secuencia persona:id_seq: %w", err)
	}

	created := 0

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

		// Obtener nuevo ID con secuencia
		newID, err := config.RedisClient.Incr(ctx, "persona:id_seq").Result()
		if err != nil {
			log.Printf("❌ Error incrementando secuencia: %v", err)
			continue
		}
		personKey := fmt.Sprintf("persona:%d", newID)
		personData := map[string]interface{}{
			"id":              fmt.Sprintf("%d", newID),
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
		err = config.RedisClient.Set(ctx, personaEmail, fmt.Sprintf("%d", newID), 0).Err()
		if err != nil {
			log.Printf("⚠️  Error creando mapeo email->userID para %s: %v", personaEmail, err)
		}

		log.Printf("✅ Creada %s - %s | Email: %s | Barrio:%d, Ruta:%d, Camión:%d",
			personKey, personaNombre, personaEmail, neighborhoodID, routeNumber, route.TruckID)

		created++
	}

	totalPersonas := created
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
// Creación y eliminación individual de personas (Redis)
// =====================

// Persona model para responses desde el servicio
type Persona struct {
	ID             string `json:"id"`
	Nombre         string `json:"nombre"`
	Email          string `json:"email"`
	NeighborhoodID int    `json:"neighborhood_id"`
	RouteNumber    int    `json:"route_number"`
	TruckID        int    `json:"truck_id"`
}

// CreatePersonaRequest datos necesarios para crear una persona
type CreatePersonaRequest struct {
	Nombre         string `json:"nombre" binding:"required,min=2"`
	Email          string `json:"email" binding:"required,email"`
	NeighborhoodID int    `json:"neighborhood_id" binding:"required,min=1"`
	RouteNumber    int    `json:"route_number" binding:"required,min=1"`
	TruckID        int    `json:"truck_id" binding:"required,min=1"`
}

// CreatePersonaRedis crea una persona en Redis y devuelve el registro creado
func CreatePersonaRedis(req CreatePersonaRequest) (*Persona, error) {
	if config.RedisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	// Evitar duplicidad por email
	if exists, _ := config.RedisClient.Exists(ctx, req.Email).Result(); exists > 0 {
		return nil, fmt.Errorf("email ya existe")
	}

	// Usar un contador secuencial para IDs
	id, err := config.RedisClient.Incr(ctx, "persona:id_seq").Result()
	if err != nil {
		return nil, fmt.Errorf("error incrementando secuencia: %v", err)
	}

	personKey := fmt.Sprintf("persona:%d", id)

	// Auto-reparación: si el ID generado ya existe (secuencia desfasada),
	// recalcular max ID actual y ajustar la secuencia
	if exists, _ := config.RedisClient.Exists(ctx, personKey).Result(); exists > 0 {
		personas, _ := config.RedisClient.LRange(ctx, "personas", 0, -1).Result()
		maxID := id
		for _, k := range personas {
			// k tiene formato "persona:<n>"
			var n int
			if _, err := fmt.Sscanf(k, "persona:%d", &n); err == nil {
				if int64(n) > maxID {
					maxID = int64(n)
				}
			}
		}
		// Setear secuencia al max y volver a INCR para tomar el siguiente libre
		if err := config.RedisClient.Set(ctx, "persona:id_seq", maxID, 0).Err(); err == nil {
			id, err = config.RedisClient.Incr(ctx, "persona:id_seq").Result()
			if err != nil {
				return nil, fmt.Errorf("error ajustando secuencia: %v", err)
			}
			personKey = fmt.Sprintf("persona:%d", id)
		}
	}

	// Guardar hash
	data := map[string]interface{}{
		"id":              strconv.FormatInt(id, 10),
		"nombre":          req.Nombre,
		"email":           req.Email,
		"neighborhood_id": strconv.Itoa(req.NeighborhoodID),
		"route_number":    strconv.Itoa(req.RouteNumber),
		"truck_id":        strconv.Itoa(req.TruckID),
		"created_at":      time.Now().Format(time.RFC3339),
	}
	if err := config.RedisClient.HSet(ctx, personKey, data).Err(); err != nil {
		return nil, fmt.Errorf("error guardando persona: %v", err)
	}

	// Añadir a la lista
	if err := config.RedisClient.LPush(ctx, "personas", personKey).Err(); err != nil {
		return nil, fmt.Errorf("error agregando a lista: %v", err)
	}

	// Mapear email -> userID
	if err := config.RedisClient.Set(ctx, req.Email, strconv.FormatInt(id, 10), 0).Err(); err != nil {
		log.Printf("warning: error creando mapeo email->userID: %v", err)
	}

	return &Persona{
		ID:             strconv.FormatInt(id, 10),
		Nombre:         req.Nombre,
		Email:          req.Email,
		NeighborhoodID: req.NeighborhoodID,
		RouteNumber:    req.RouteNumber,
		TruckID:        req.TruckID,
	}, nil
}

// DeletePersonaRedis elimina una persona por ID (hash, lista y mapeo de email)
func DeletePersonaRedis(id int) error {
	if config.RedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	if id <= 0 {
		return fmt.Errorf("id inválido")
	}

	personKey := fmt.Sprintf("persona:%d", id)

	// Verificar existencia y obtener email para limpiar el mapeo
	exists, err := config.RedisClient.Exists(ctx, personKey).Result()
	if err != nil {
		return fmt.Errorf("error verificando existencia: %v", err)
	}
	if exists == 0 {
		return fmt.Errorf("persona %d no encontrada", id)
	}

	// Obtener email antes de eliminar
	email, _ := config.RedisClient.HGet(ctx, personKey, "email").Result()

	if err := config.RedisClient.Del(ctx, personKey).Err(); err != nil {
		return fmt.Errorf("error eliminando persona: %v", err)
	}

	// Remover de la lista
	if err := config.RedisClient.LRem(ctx, "personas", 0, personKey).Err(); err != nil {
		return fmt.Errorf("error removiendo de lista: %v", err)
	}

	// Eliminar mapeo de email si existía
	if email != "" {
		_ = config.RedisClient.Del(ctx, email).Err()
	}

	return nil
}
