package services

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Point struct {
	ID  int     `json:"id"`
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// CreateTachoRequest representa la estructura de datos para crear un tacho
type CreateTachoRequest struct {
	// Datos para MySQL
	IdTipo    int     `json:"id_tipo"`
	IdEstado  int     `json:"id_estado"`
	Capacidad float64 `json:"capacidad"`

	// Datos para Neo4j
	Barrio    string  `json:"barrio"`
	Direccion string  `json:"direccion"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Prioridad int     `json:"prioridad"`
}

// CreateTachoResponse representa la respuesta al crear un tacho
type CreateTachoResponse struct {
	Message   string `json:"message"`
	TachoID   int    `json:"tacho_id"`
	NeoNodeID string `json:"neo_node_id"`
}

// Haversine calculates the distance between two coordinates
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Radius of the Earth en km
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	lat1 = lat1 * math.Pi / 180.0
	lat2 = lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// GetDistances gets the 'tachos' in an 'zona' and sorts them by distance
func GetDistances(zonaID int) ([]Point, error) {
	driver, err := config.ConnectNeo()
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar a Neo4j: %v", err)
	}
	defer driver.Close(context.Background())

	session := driver.NewSession(context.Background(), neo4j.SessionConfig{
		DatabaseName: os.Getenv("NEO4J_DATABASE"),
	})
	defer session.Close(context.Background())

	// Mapping zonaID -> barrio
	zonaToBarrio := map[int]string{
		1: "CHACARITA",
		2: "MONTE CASTRO",
		3: "BOEDO",
		4: "VILLA CRESPO",
	}

	barrio, ok := zonaToBarrio[zonaID]
	if !ok {
		return nil, fmt.Errorf("zonaID desconocido")
	}

	// Query to fetch Neo4j nodes by barrio
	query := `
	MATCH (t:Tacho)
	WHERE t.barrio = $barrio
	RETURN t.id AS id, t.location.latitude AS lat, t.location.longitude AS lng
	`

	result, err := session.ExecuteRead(context.Background(), func(tx neo4j.ManagedTransaction) (any, error) {
		records, err := tx.Run(context.Background(), query, map[string]any{
			"barrio": barrio,
		})
		if err != nil {
			return nil, err
		}

		points := []Point{}
		idCounter := 1

		for records.Next(context.Background()) {
			rec := records.Record()
			latVal, _ := rec.Get("lat")
			lngVal, _ := rec.Get("lng")

			lat, ok1 := latVal.(float64)
			lng, ok2 := lngVal.(float64)
			if !ok1 || !ok2 {
				continue
			}

			points = append(points, Point{
				ID:  idCounter,
				Lat: lat,
				Lng: lng,
			})
			idCounter++
		}

		return points, records.Err()
	})
	if err != nil {
		return nil, err
	}

	points := result.([]Point)

	// Order by distance from the first using sort.Slice
	if len(points) > 1 {
		ref := points[0]
		sort.Slice(points[1:], func(i, j int) bool {
			distI := haversine(ref.Lat, ref.Lng, points[i+1].Lat, points[i+1].Lng)
			distJ := haversine(ref.Lat, ref.Lng, points[j+1].Lat, points[j+1].Lng)
			return distI < distJ
		})
	}

	return points, nil
}

// GetUserByEmail obtiene el número de persona asociado a un email desde Redis
func GetUserByEmail(email string) (string, error) {
	if config.RedisClient == nil {
		return "", fmt.Errorf("redis client not available")
	}
	config.LoadDummyUsers()
	// Buscar el valor asociado al email en Redis
	result, err := config.RedisClient.Get(context.Background(), email).Result()
	if err != nil {
		return "", fmt.Errorf("email no encontrado en Redis: %v", err)
	}

	return result, nil
}

// GetPersonaByKey obtiene los datos de una persona desde Redis usando su clave
func GetPersonaByKey(personaKey string) (map[string]interface{}, error) {
	if config.RedisClient == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	// Obtener todos los campos del hash de la persona
	result, err := config.RedisClient.HGetAll(context.Background(), personaKey).Result()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo persona: %v", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("persona no encontrada")
	}

	// Convertir map[string]string a map[string]interface{}
	persona := make(map[string]interface{})
	for k, v := range result {
		persona[k] = v
	}

	return persona, nil
}

// CreateTacho crea un tacho en MySQL y Neo4j
func CreateTacho(request CreateTachoRequest) (*CreateTachoResponse, error) {
	// Generar el ID personalizado (direccion|barrio)
	customID := fmt.Sprintf("%s|%s", request.Direccion, request.Barrio)

	// Primero crear en Neo4j para obtener el ID del nodo
	neoNodeID, err := createTachoInNeo4j(request)
	if err != nil {
		return nil, fmt.Errorf("error creando tacho en Neo4j: %v", err)
	}

	// Luego crear en MySQL usando el ID personalizado en lugar del ID interno de Neo4j
	tachoID, err := createTachoInMySQL(request, customID)
	if err != nil {
		// Si falla MySQL, intentar limpiar Neo4j (rollback)
		// TODO: Implementar rollback en Neo4j si es necesario
		return nil, fmt.Errorf("error creando tacho en MySQL: %v", err)
	}

	return &CreateTachoResponse{
		Message:   "Tacho creado exitosamente",
		TachoID:   tachoID,
		NeoNodeID: neoNodeID,
	}, nil
}

// DeleteTacho elimina un tacho de MySQL y Neo4j usando el ID personalizado
func DeleteTacho(customID string) error {
	var errorsFound []string

	// Intentar eliminar de MySQL
	err := deleteTachoFromMySQL(0, customID)
	if err != nil {
		errorsFound = append(errorsFound, fmt.Sprintf("MySQL: %v", err))
	}

	// Intentar eliminar de Neo4j (siempre, sin importar si falló MySQL)
	err = deleteTachoFromNeo4j("", customID)
	if err != nil {
		errorsFound = append(errorsFound, fmt.Sprintf("Neo4j: %v", err))
	}

	// Si ambos fallaron, retornar error
	if len(errorsFound) == 2 {
		return fmt.Errorf("no se pudo eliminar de ninguna base de datos: %s", strings.Join(errorsFound, "; "))
	}

	// Si al menos uno funcionó, es éxito (aunque logueamos warnings)
	if len(errorsFound) == 1 {
		fmt.Printf("Warning durante eliminación: %s\n", errorsFound[0])
	}

	return nil
}
