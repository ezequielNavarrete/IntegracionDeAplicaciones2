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

// BinMongo representa un tacho en MongoDB
type BinMongo struct {
	ID           int     `bson:"id" json:"id"`
	Lat          float64 `bson:"lat" json:"lat"`
	Lon          float64 `bson:"lon" json:"lon"`
	Neighborhood int     `bson:"neighborhood" json:"neighborhood"`
}

// GetAllTachosFromMongoDB obtiene todos los tachos desde MongoDB
func GetAllTachosFromMongoDB() (map[int]BinMongo, error) {
	collection, err := config.GetMongoCollection("bins")
	if err != nil {
		return nil, fmt.Errorf("error getting MongoDB collection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("error finding bins: %w", err)
	}
	defer cursor.Close(ctx)

	binsMap := make(map[int]BinMongo)
	for cursor.Next(ctx) {
		var bin BinMongo
		if err := cursor.Decode(&bin); err != nil {
			continue // Skip invalid documents
		}
		binsMap[bin.ID] = bin
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return binsMap, nil
}

// createTachoInMongoDB crea un tacho en MongoDB
func createTachoInMongoDB(request CreateTachoRequest) error {
	collection, err := config.GetMongoCollection("bins")
	if err != nil {
		return fmt.Errorf("error getting MongoDB collection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Crear documento para MongoDB
	binDoc := bson.M{
		"id":           request.MongoID,
		"lat":          request.Latitude,
		"lon":          request.Longitude,
		"neighborhood": request.Neighborhood,
	}

	_, err = collection.InsertOne(ctx, binDoc)
	if err != nil {
		return fmt.Errorf("error inserting bin: %w", err)
	}

	return nil
}

// deleteTachoFromMongoDB elimina un tacho de MongoDB por su ID
func deleteTachoFromMongoDB(mongoID int) error {
	collection, err := config.GetMongoCollection("bins")
	if err != nil {
		return fmt.Errorf("error getting MongoDB collection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"id": mongoID})
	if err != nil {
		return fmt.Errorf("error deleting bin: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("no se encontró el tacho con id %d", mongoID)
	}

	return nil
}

// DepotMongo representa un depot (centro) en MongoDB
type DepotMongo struct {
	ID           int     `bson:"id" json:"id"`
	Lat          float64 `bson:"lat" json:"lat"`
	Lon          float64 `bson:"lon" json:"lon"`
	Neighborhood int     `bson:"neighborhood" json:"neighborhood"`
}

// GetAllDepotsFromMongoDB obtiene todos los depots desde MongoDB
func GetAllDepotsFromMongoDB() (map[int]DepotMongo, error) {
	collection, err := config.GetMongoCollection("depot")
	if err != nil {
		return nil, fmt.Errorf("error getting MongoDB collection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("error finding depots: %w", err)
	}
	defer cursor.Close(ctx)

	depotsMap := make(map[int]DepotMongo)
	for cursor.Next(ctx) {
		var depot DepotMongo
		if err := cursor.Decode(&depot); err != nil {
			continue // Skip invalid documents
		}
		depotsMap[depot.ID] = depot
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return depotsMap, nil
}

// GetDepotByIDFromMongoDB obtiene un depot específico desde MongoDB por su ID
func GetDepotByIDFromMongoDB(depotID int) (*DepotMongo, error) {
	collection, err := config.GetMongoCollection("depot")
	if err != nil {
		return nil, fmt.Errorf("error getting MongoDB collection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var depot DepotMongo
	err = collection.FindOne(ctx, bson.M{"id": depotID}).Decode(&depot)
	if err != nil {
		return nil, fmt.Errorf("error finding depot: %w", err)
	}

	return &depot, nil
}
