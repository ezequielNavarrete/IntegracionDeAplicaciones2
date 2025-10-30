package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

// InitMongoDB inicializa la conexión a MongoDB
func InitMongoDB() error {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		return fmt.Errorf("MONGODB_URI environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("error connecting to MongoDB: %w", err)
	}

	// Verificar la conexión
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("error pinging MongoDB: %w", err)
	}

	MongoClient = client
	fmt.Println("✓ Connected to MongoDB successfully")
	return nil
}

// GetMongoDatabase obtiene la base de datos de MongoDB
func GetMongoDatabase() (*mongo.Database, error) {
	if MongoClient == nil {
		return nil, fmt.Errorf("MongoDB client not initialized")
	}

	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "tachos" // Default
	}

	return MongoClient.Database(dbName), nil
}

// GetMongoCollection obtiene una colección específica
func GetMongoCollection(collectionName string) (*mongo.Collection, error) {
	db, err := GetMongoDatabase()
	if err != nil {
		return nil, err
	}

	return db.Collection(collectionName), nil
}

// CloseMongoDB cierra la conexión a MongoDB
func CloseMongoDB() error {
	if MongoClient == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return MongoClient.Disconnect(ctx)
}
