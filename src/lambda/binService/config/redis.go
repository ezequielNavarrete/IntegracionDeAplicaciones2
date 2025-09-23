package config

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var ctx = context.Background()

// CamionOperativo representa un camión operativo desde MySQL
type CamionOperativo struct {
	ID     int `json:"id_camion" gorm:"column:id_camion"`
	Estado int `json:"id_estado" gorm:"column:id_estado"`
	Tipo   int `json:"id_tipo" gorm:"column:id_tipo"`
}

// ZonaDisponible representa una zona desde MySQL
type ZonaDisponible struct {
	ID     int    `json:"id_zona" gorm:"column:id_zona"`
	Nombre string `json:"nombre" gorm:"column:nombre"`
}

func ConnectRedis() {
	// Configurar Redis client
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
		log.Printf("⚠️  REDIS_URL no encontrado, usando default: %s", redisURL)
	} else {
		log.Printf("🔧 Redis Config: %s", redisURL)
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Error parsing Redis URL: %v", err)
	}

	RedisClient = redis.NewClient(opt)

	// Test connection
	_, err = RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	log.Println("Redis connected successfully")

	// Inicializar datos si no existen
	InitializePersonsData()
	LoadDummyUsers()
}

func InitializePersonsData() {
	// Verificar si ya existe la lista de personas
	exists, err := RedisClient.Exists(ctx, "personas").Result()
	if err != nil {
		log.Printf("Error checking if personas exists: %v", err)
		return
	}

	if exists > 0 {
		log.Println("Personas data already exists, skipping initialization")
		return
	}

	log.Println("Initializing personas data from MySQL...")

	// Obtener camiones operativos desde MySQL
	camionesOperativos, err := getCamionesOperativos()
	if err != nil {
		log.Printf("Error getting operational trucks: %v", err)
		return
	}

	if len(camionesOperativos) == 0 {
		log.Println("No operational trucks found, cannot initialize personas")
		return
	}

	// Obtener zonas disponibles desde MySQL
	zonasDisponibles, err := getZonasDisponibles()
	if err != nil {
		log.Printf("Error getting available zones: %v", err)
		return
	}

	if len(zonasDisponibles) == 0 {
		log.Println("No zones found, cannot initialize personas")
		return
	}

	// Mezclar zonas para asignación aleatoria sin repetir
	rand.Seed(time.Now().UnixNano())
	zonasShuffled := make([]ZonaDisponible, len(zonasDisponibles))
	copy(zonasShuffled, zonasDisponibles)

	for i := len(zonasShuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		zonasShuffled[i], zonasShuffled[j] = zonasShuffled[j], zonasShuffled[i]
	}

	// Determinar cuántas personas crear (mínimo entre 10 y zonas disponibles)
	maxPersonas := 10
	if len(zonasDisponibles) < maxPersonas {
		maxPersonas = len(zonasDisponibles)
	}

	log.Printf("Creating %d personas with %d operational trucks and %d zones",
		maxPersonas, len(camionesOperativos), len(zonasDisponibles))

	// Crear personas
	for i := 1; i <= maxPersonas; i++ {
		personKey := fmt.Sprintf("persona:%d", i)

		// Asignar zona sin repetir (tomar de la lista mezclada)
		zona := zonasShuffled[i-1]

		// Asignar camión aleatoriamente de los operativos
		camionIndex := rand.Intn(len(camionesOperativos))
		camion := camionesOperativos[camionIndex]

		// Crear hash para la persona
		personData := map[string]interface{}{
			"id":          fmt.Sprintf("%d", i),
			"zona_id":     fmt.Sprintf("%d", zona.ID),
			"zona_nombre": zona.Nombre,
			"camion_id":   fmt.Sprintf("%d", camion.ID),
			"camion_tipo": fmt.Sprintf("%d", camion.Tipo),
			"nombre":      fmt.Sprintf("Persona_%d", i),
			"estado":      "activo",
			"created_at":  time.Now().Format(time.RFC3339),
		}

		// Guardar hash en Redis
		err := RedisClient.HMSet(ctx, personKey, personData).Err()
		if err != nil {
			log.Printf("Error creating hash for %s: %v", personKey, err)
			continue
		}

		// Agregar a la lista de personas
		err = RedisClient.LPush(ctx, "personas", personKey).Err()
		if err != nil {
			log.Printf("Error adding %s to personas list: %v", personKey, err)
			continue
		}

		log.Printf("Created %s - Zona: %s (ID:%d), Camión: %d (Tipo:%d)",
			personKey, zona.Nombre, zona.ID, camion.ID, camion.Tipo)
	}

	// Obtener y mostrar estadísticas
	listLength, _ := RedisClient.LLen(ctx, "personas").Result()
	log.Printf("Initialization completed. Total personas: %d", listLength)

	// Mostrar resumen de asignaciones
	showPersonsSummary()
}

func showPersonsSummary() {
	log.Println("=== RESUMEN DE PERSONAS CREADAS ===")

	// Obtener todas las personas de la lista
	personas, err := RedisClient.LRange(ctx, "personas", 0, -1).Result()
	if err != nil {
		log.Printf("Error getting personas list: %v", err)
		return
	}

	for _, personKey := range personas {
		// Obtener datos de cada persona
		_, err := RedisClient.HGetAll(ctx, personKey).Result()
		if err != nil {
			log.Printf("Error getting data for %s: %v", personKey, err)
			continue
		}
	}

	log.Println("====================================")
}

// getCamionesOperativos obtiene camiones operativos con JOIN desde MySQL usando GORM
func getCamionesOperativos() ([]CamionOperativo, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var camiones []CamionOperativo

	// Usar Raw query con GORM para el JOIN
	result := DB.Raw(`
		SELECT c.id_camion, c.id_estado, c.id_tipo 
		FROM Camiones c
		WHERE c.id_estado = 1
	`).Scan(&camiones)

	if result.Error != nil {
		return nil, fmt.Errorf("error querying operational trucks: %v", result.Error)
	}

	log.Printf("Found %d operational trucks", len(camiones))

	// Debug: mostrar los primeros camiones encontrados
	for i, camion := range camiones {
		if i < 3 { // Solo mostrar los primeros 3 para no saturar logs
			log.Printf("🚚 Camion %d: ID=%d, Estado=%d, Tipo=%d", i+1, camion.ID, camion.Estado, camion.Tipo)
		}
	}

	return camiones, nil
}

// getZonasDisponibles obtiene todas las zonas desde MySQL usando GORM
func getZonasDisponibles() ([]ZonaDisponible, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var zonas []ZonaDisponible

	// Usar Raw query con GORM
	result := DB.Raw("SELECT id_zona, nombre FROM Zona ORDER BY id_zona").Scan(&zonas)

	if result.Error != nil {
		return nil, fmt.Errorf("error querying zones: %v", result.Error)
	}

	log.Printf("Found %d available zones", len(zonas))

	// Debug: mostrar las primeras zonas encontradas
	for i, zona := range zonas {
		if i < 3 { // Solo mostrar las primeras 3 para no saturar logs
			log.Printf("🏭 Zona %d: ID=%d, Nombre='%s'", i+1, zona.ID, zona.Nombre)
		}
	}

	return zonas, nil
}

// LoadDummyUsers carga 10 pares email->usuario en Redis
func LoadDummyUsers() {
	flagKey := "dummy_users_loaded"

	// Verifico si ya existe la flag
	val, err := RedisClient.Get(ctx, flagKey).Result()
	if err == nil && val == "true" {
		fmt.Println("Usuarios dummy ya estaban cargados, no se hace nada ✅")
		return
	}
	users := map[string]string{
		"eze@example.com":    "1",
		"leo@example.com":    "2",
		"juan@example.com":   "3",
		"sergio@example.com": "4",
		"fede@example.com":   "5",
		"gonza@example.com":  "6",
		"maria@example.com":  "7",
		"lucia@example.com":  "8",
		"sofia@example.com":  "9",
		"camila@example.com": "10",
	}

	for email, name := range users {
		err := RedisClient.Set(ctx, email, name, 0).Err()
		if err != nil {
			fmt.Printf("Error cargando %s: %v\n", email, err)
		}
	}

	err = RedisClient.Set(ctx, flagKey, "true", 0).Err()
	if err != nil {
		fmt.Printf("Error seteando flag: %v\n", err)
	}

	fmt.Println("Datos dummy cargados en Redis ✅")
}
