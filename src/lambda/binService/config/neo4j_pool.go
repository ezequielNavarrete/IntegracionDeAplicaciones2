package config

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	neo4jDriver neo4j.DriverWithContext
	neo4jOnce   sync.Once
	neo4jError  error
)

// GetNeo4jDriver obtiene o crea el driver de Neo4j (singleton)
func GetNeo4jDriver() (neo4j.DriverWithContext, error) {
	neo4jOnce.Do(func() {
		neo4jDriver, neo4jError = ConnectNeo()
	})
	return neo4jDriver, neo4jError
}

// GetNeo4jSession obtiene una nueva sesión reutilizando el driver
func GetNeo4jSession() (neo4j.SessionWithContext, error) {
	driver, err := GetNeo4jDriver()
	if err != nil {
		return nil, fmt.Errorf("error getting Neo4j driver: %v", err)
	}

	session := driver.NewSession(context.Background(), neo4j.SessionConfig{
		DatabaseName: os.Getenv("NEO4J_DATABASE"),
	})

	return session, nil
}

// CloseNeo4jDriver cierra el driver global (llamar al cerrar la app)
func CloseNeo4jDriver() {
	if neo4jDriver != nil {
		neo4jDriver.Close(context.Background())
	}
}
