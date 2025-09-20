package services

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/demo/config"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Point struct {
	ID  int     `json:"id"`
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
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
