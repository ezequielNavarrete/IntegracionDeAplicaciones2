package services

import (
	"context"
	"fmt"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// getSession obtiene una sesión reutilizable de Neo4j usando el pool
func getSession() (neo4j.SessionWithContext, error) {
	session, err := config.GetNeo4jSession()
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener sesión de Neo4j: %v", err)
	}
	return session, nil
}

// createTachoInNeo4j crea un nodo Tacho en Neo4j y retorna su ID
func createTachoInNeo4j(request CreateTachoRequest) (string, error) {
	session, err := getSession()
	if err != nil {
		return "", err
	}
	defer session.Close(context.Background())

	// Crear el nodo con point() para la ubicación geográfica
	query := `
		CREATE (t:Tacho {
			barrio: $barrio,
			direccion: $direccion,
			id: $id,
			location: point({latitude: $latitude, longitude: $longitude}),
			prioridad: $prioridad
		})
		RETURN elementId(t) as nodeId
	`

	// Generar un ID único para el tacho (puedes usar una estrategia diferente)
	tachoNeoID := fmt.Sprintf("%s|%s", request.Direccion, request.Barrio)

	result, err := session.ExecuteWrite(context.Background(), func(tx neo4j.ManagedTransaction) (interface{}, error) {
		ctx := context.Background()
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"barrio":    request.Barrio,
			"direccion": request.Direccion,
			"id":        tachoNeoID,
			"latitude":  request.Latitude,
			"longitude": request.Longitude,
			"prioridad": request.Prioridad,
		})
		if err != nil {
			return nil, err
		}

		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}

		nodeId, _ := record.Get("nodeId")
		return nodeId.(string), nil
	})

	if err != nil {
		return "", err
	}

	return result.(string), nil
}

// deleteTachoFromNeo4j elimina un nodo Tacho de Neo4j por su elementId o por su id personalizado
func deleteTachoFromNeo4j(neoNodeID string, customID string) error {
	session, err := getSession()
	if err != nil {
		return err
	}
	defer session.Close(context.Background())

	var query string
	var params map[string]interface{}

	if neoNodeID != "" {
		// Eliminar por elementId (ID interno de Neo4j)
		query = `
			MATCH (t:Tacho)
			WHERE elementId(t) = $nodeId
			DELETE t
			RETURN count(t) as deletedCount
		`
		params = map[string]interface{}{
			"nodeId": neoNodeID,
		}
	} else if customID != "" {
		// Eliminar por ID personalizado (direccion|barrio)
		query = `
			MATCH (t:Tacho {id: $customId})
			DELETE t
			RETURN count(t) as deletedCount
		`
		params = map[string]interface{}{
			"customId": customID,
		}
	} else {
		return fmt.Errorf("debe proporcionar neoNodeID o customID para eliminar")
	}

	result, err := session.ExecuteWrite(context.Background(), func(tx neo4j.ManagedTransaction) (interface{}, error) {
		ctx := context.Background()
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}

		deletedCount, _ := record.Get("deletedCount")
		return deletedCount, nil
	})

	if err != nil {
		return fmt.Errorf("error deleting tacho from Neo4j: %v", err)
	}

	deletedCount := result.(int64)
	if deletedCount == 0 {
		return fmt.Errorf("no se encontró el tacho en Neo4j para eliminar")
	}

	return nil
}

// getTachoFromNeo4j obtiene un tacho de Neo4j por su elementId o ID personalizado
func getTachoFromNeo4j(neoNodeID string, customID string) (*TachoNeo4j, error) {
	session, err := getSession()
	if err != nil {
		return nil, err
	}
	defer session.Close(context.Background())

	var query string
	var params map[string]interface{}

	if neoNodeID != "" {
		query = `
			MATCH (t:Tacho)
			WHERE elementId(t) = $nodeId
			RETURN t.barrio as barrio, t.direccion as direccion, t.id as id, 
				   t.location.latitude as latitude, t.location.longitude as longitude, 
				   t.prioridad as prioridad, elementId(t) as nodeId
		`
		params = map[string]interface{}{"nodeId": neoNodeID}
	} else if customID != "" {
		query = `
			MATCH (t:Tacho {id: $customId})
			RETURN t.barrio as barrio, t.direccion as direccion, t.id as id,
				   t.location.latitude as latitude, t.location.longitude as longitude,
				   t.prioridad as prioridad, elementId(t) as nodeId
		`
		params = map[string]interface{}{"customId": customID}
	} else {
		return nil, fmt.Errorf("debe proporcionar neoNodeID o customID")
	}

	result, err := session.ExecuteRead(context.Background(), func(tx neo4j.ManagedTransaction) (interface{}, error) {
		ctx := context.Background()
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}

		barrio, _ := record.Get("barrio")
		direccion, _ := record.Get("direccion")
		id, _ := record.Get("id")
		latitude, _ := record.Get("latitude")
		longitude, _ := record.Get("longitude")
		prioridad, _ := record.Get("prioridad")
		nodeId, _ := record.Get("nodeId")

		return &TachoNeo4j{
			NodeID:    nodeId.(string),
			Barrio:    barrio.(string),
			Direccion: direccion.(string),
			ID:        id.(string),
			Latitude:  latitude.(float64),
			Longitude: longitude.(float64),
			Prioridad: int(prioridad.(int64)),
		}, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error getting tacho from Neo4j: %v", err)
	}

	return result.(*TachoNeo4j), nil
}

// TachoNeo4j representa un tacho en Neo4j
type TachoNeo4j struct {
	NodeID    string  `json:"node_id"`
	Barrio    string  `json:"barrio"`
	Direccion string  `json:"direccion"`
	ID        string  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Prioridad int     `json:"prioridad"`
}

// GetAllTachosCoordinates obtiene las coordenadas de todos los tachos desde Neo4j
func GetAllTachosCoordinates() (map[string]TachoNeo4j, error) {
	session, err := getSession()
	if err != nil {
		return nil, err
	}
	defer session.Close(context.Background())

	query := `
		MATCH (t:Tacho)
		RETURN t.id as id, t.barrio as barrio, t.direccion as direccion,
			   t.location.latitude as latitude, t.location.longitude as longitude,
			   t.prioridad as prioridad
	`

	result, err := session.ExecuteRead(context.Background(), func(tx neo4j.ManagedTransaction) (interface{}, error) {
		ctx := context.Background()
		records, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		coordsMap := make(map[string]TachoNeo4j)
		for records.Next(ctx) {
			record := records.Record()
			
			id, _ := record.Get("id")
			barrio, _ := record.Get("barrio")
			direccion, _ := record.Get("direccion")
			latitude, _ := record.Get("latitude")
			longitude, _ := record.Get("longitude")
			prioridad, _ := record.Get("prioridad")

			tachoNeo := TachoNeo4j{
				ID:        id.(string),
				Barrio:    barrio.(string),
				Direccion: direccion.(string),
				Latitude:  latitude.(float64),
				Longitude: longitude.(float64),
				Prioridad: int(prioridad.(int64)),
			}

			coordsMap[id.(string)] = tachoNeo
		}

		return coordsMap, records.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("error getting coordinates from Neo4j: %v", err)
	}

	return result.(map[string]TachoNeo4j), nil
}
