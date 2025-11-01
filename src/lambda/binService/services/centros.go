package services

import (
	"fmt"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
)

// Estructura para representar un centro con información completa (MySQL + MongoDB)
type Centro struct {
	IDCentro     int     `json:"id_centro" gorm:"column:id_centro"`
	NombreTipo   string  `json:"nombre_tipo" gorm:"column:nombre_tipo"`
	IDMongo      int     `json:"id_mongo" gorm:"column:id_mongo"`
	Neighborhood int     `json:"neighborhood"`
	Longitud     float64 `json:"longitud"`
	Latitud      float64 `json:"latitud"`
}

// Estructura para respuesta de centros
type CentrosResponse struct {
	Centros []Centro `json:"centros"`
	Total   int      `json:"total"`
}

// Estructura para respuesta de un centro individual
type CentroResponse struct {
	Centro Centro `json:"centro"`
}

// GetAllCentros obtiene todos los centros con información de tipo (MySQL) y datos adicionales (MongoDB)
func GetAllCentros() (*CentrosResponse, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Estructura temporal para datos de MySQL
	type CentroMySQL struct {
		IDCentro   int    `gorm:"column:id_centro"`
		IDTipo     int    `gorm:"column:id_tipo"`
		NombreTipo string `gorm:"column:nombre_tipo"`
		IDMongo    int    `gorm:"column:id_mongo"`
	}

	var centrosMySQL []CentroMySQL

	// Query con JOIN para obtener información de MySQL
	query := `
		SELECT 
			c.id_centro,
			c.id_tipo,
			tc.nombre_tipo,
			c.id_mongo
		FROM Centro c
		LEFT JOIN Tipo_centro tc ON c.id_tipo = tc.id_tipo
		ORDER BY c.id_centro ASC
	`

	if err := config.DB.Raw(query).Scan(&centrosMySQL).Error(); err != nil {
		return nil, fmt.Errorf("error querying centros from MySQL: %v", err)
	}

	// Obtener depots desde MongoDB
	depotsMap, err := GetAllDepotsFromMongoDB()
	if err != nil {
		return nil, fmt.Errorf("error getting depots from MongoDB: %v", err)
	}

	// Convertir a estructura completa y obtener datos de MongoDB
	var centros []Centro
	for _, centroMySQL := range centrosMySQL {
		centro := Centro{
			IDCentro:   centroMySQL.IDCentro,
			NombreTipo: centroMySQL.NombreTipo,
			IDMongo:    centroMySQL.IDMongo,
		}

		// Obtener información adicional de MongoDB si existe IDMongo
		if depotData, found := depotsMap[centroMySQL.IDMongo]; found {
			centro.Neighborhood = depotData.Neighborhood
			centro.Longitud = depotData.Lon
			centro.Latitud = depotData.Lat
		}

		centros = append(centros, centro)
	}

	return &CentrosResponse{
		Centros: centros,
		Total:   len(centros),
	}, nil
}

// GetCentroByID obtiene un centro específico por ID con información completa
func GetCentroByID(centroID int) (*CentroResponse, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Estructura temporal para datos de MySQL
	type CentroMySQL struct {
		IDCentro   int    `gorm:"column:id_centro"`
		IDTipo     int    `gorm:"column:id_tipo"`
		NombreTipo string `gorm:"column:nombre_tipo"`
		IDMongo    int    `gorm:"column:id_mongo"`
	}

	var centroMySQL CentroMySQL

	// Query con JOIN para obtener información de MySQL
	query := `
		SELECT 
			c.id_centro,
			c.id_tipo,
			tc.nombre_tipo,
			c.id_mongo
		FROM Centro c
		LEFT JOIN Tipo_centro tc ON c.id_tipo = tc.id_tipo
		WHERE c.id_centro = ?
	`

	if err := config.DB.Raw(query, centroID).Scan(&centroMySQL).Error(); err != nil {
		return nil, fmt.Errorf("error querying centro from MySQL: %v", err)
	}

	// Verificar si se encontró el centro
	if centroMySQL.IDCentro == 0 {
		return nil, fmt.Errorf("centro with ID %d not found", centroID)
	}

	// Crear estructura completa
	centro := Centro{
		IDCentro:   centroMySQL.IDCentro,
		NombreTipo: centroMySQL.NombreTipo,
		IDMongo:    centroMySQL.IDMongo,
	}

	// Obtener información adicional de MongoDB si existe IDMongo
	if centroMySQL.IDMongo > 0 {
		depotData, err := GetDepotByIDFromMongoDB(centroMySQL.IDMongo)
		if err != nil {
			return nil, fmt.Errorf("error getting MongoDB data for centro %d: %v", centroMySQL.IDMongo, err)
		}
		centro.Neighborhood = depotData.Neighborhood
		centro.Longitud = depotData.Lon
		centro.Latitud = depotData.Lat
	}

	return &CentroResponse{
		Centro: centro,
	}, nil
}

// DEPRECATED: Las funciones de Neo4j ya no se usan para centros
// Los centros ahora se obtienen desde MongoDB (colección depot)
