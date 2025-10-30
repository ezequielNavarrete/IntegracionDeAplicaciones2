package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// NeighborhoodStats representa las estadísticas de tachos por barrio
type NeighborhoodStats struct {
	Neighborhood int `bson:"_id" json:"neighborhood"`
	TotalBins    int `bson:"total_tachos" json:"total_bins"`
}

// GetBinsStatsByNeighborhood obtiene la cantidad de tachos por barrio
func GetBinsStatsByNeighborhood() ([]NeighborhoodStats, error) {
	collection, err := config.GetMongoCollection("bins")
	if err != nil {
		return nil, fmt.Errorf("error getting MongoDB collection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Agregación para contar tachos por barrio
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{
			"_id":          "$neighborhood",
			"total_tachos": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error executing aggregation: %w", err)
	}
	defer cursor.Close(ctx)

	var stats []NeighborhoodStats
	if err := cursor.All(ctx, &stats); err != nil {
		return nil, fmt.Errorf("error decoding results: %w", err)
	}

	return stats, nil
}

// GetNeighborhoodsWithBins obtiene lista de barrios que tienen tachos
func GetNeighborhoodsWithBins() ([]int, error) {
	stats, err := GetBinsStatsByNeighborhood()
	if err != nil {
		return nil, err
	}

	neighborhoods := make([]int, 0, len(stats))
	for _, stat := range stats {
		neighborhoods = append(neighborhoods, stat.Neighborhood)
	}

	return neighborhoods, nil
}
