package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
)

type Point struct {
	ID  int     `json:"id"`
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// CreateTachoRequest representa la estructura de datos para crear un tacho
type CreateTachoRequest struct {
	// Datos para MySQL
	IdTipo    int     `json:"id_tipo" binding:"required"`
	IdEstado  int     `json:"id_estado" binding:"required"`
	Capacidad float64 `json:"capacidad" binding:"required"`

	// Datos para MongoDB (lat, lon, neighborhood)
	Neighborhood int     `json:"neighborhood" binding:"required"` // Número de barrio
	Latitude     float64 `json:"lat" binding:"required"`
	Longitude    float64 `json:"lon" binding:"required"`
}

// CreateTachoResponse representa la respuesta al crear un tacho
type CreateTachoResponse struct {
	Message string `json:"message"`
	TachoID int    `json:"tacho_id"`
	MongoID int    `json:"mongo_id"`
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

// GetDistances gets the 'tachos' in a neighborhood (barrio) and sorts them by distance
// DEPRECATED: Esta función usa Neo4j para tachos, que ahora están en MongoDB
// Considerar reescribir para usar MongoDB directamente
func GetDistances(zonaID int) ([]Point, error) {
	// Mapping zonaID -> neighborhood
	zonaToNeighborhood := map[int]int{
		1: 1, // CHACARITA
		2: 2, // MONTE CASTRO
		3: 3, // BOEDO
		4: 4, // VILLA CRESPO
	}

	neighborhood, ok := zonaToNeighborhood[zonaID]
	if !ok {
		return nil, fmt.Errorf("zonaID desconocido")
	}

	// Obtener tachos desde MongoDB
	binsMap, err := GetAllTachosFromMongoDB()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo tachos desde MongoDB: %v", err)
	}

	points := []Point{}
	idCounter := 1

	// Filtrar por neighborhood y convertir a Points
	for _, bin := range binsMap {
		if bin.Neighborhood == neighborhood {
			points = append(points, Point{
				ID:  idCounter,
				Lat: bin.Lat,
				Lng: bin.Lon,
			})
			idCounter++
		}
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("no se encontraron tachos en el barrio %d", neighborhood)
	}

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

// CreateTacho crea un tacho en MySQL y MongoDB
func CreateTacho(request CreateTachoRequest) (*CreateTachoResponse, error) {
	// Primero crear en MongoDB y obtener el ID generado
	mongoID, err := createTachoInMongoDB(request)
	if err != nil {
		return nil, fmt.Errorf("error creando tacho en MongoDB: %v", err)
	}

	// Luego crear en MySQL usando el MongoID
	tachoID, err := createTachoInMySQL(request, mongoID)
	if err != nil {
		// Si falla MySQL, intentar limpiar MongoDB (rollback)
		_ = deleteTachoFromMongoDB(mongoID)
		return nil, fmt.Errorf("error creando tacho en MySQL: %v", err)
	}

	return &CreateTachoResponse{
		Message: "Tacho creado exitosamente",
		TachoID: tachoID,
		MongoID: mongoID,
	}, nil
}

// DeleteTacho elimina un tacho de MySQL y MongoDB usando el ID de MongoDB
func DeleteTacho(mongoID int) error {
	var errorsFound []string

	// Intentar eliminar de MySQL
	err := deleteTachoFromMySQL(0, mongoID)
	if err != nil {
		errorsFound = append(errorsFound, fmt.Sprintf("MySQL: %v", err))
	}

	// Intentar eliminar de MongoDB
	err = deleteTachoFromMongoDB(mongoID)
	if err != nil {
		errorsFound = append(errorsFound, fmt.Sprintf("MongoDB: %v", err))
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
