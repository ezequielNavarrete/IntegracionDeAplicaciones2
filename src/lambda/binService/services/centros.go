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

// CreateCentroRequest representa la solicitud para crear un centro
type CreateCentroRequest struct {
	IdTipo       int     `json:"id_tipo" binding:"required"`
	Latitud      float64 `json:"latitud" binding:"required"`
	Longitud     float64 `json:"longitud" binding:"required"`
	Neighborhood int     `json:"neighborhood" binding:"required"`
}

// CreateCentroResponse representa la respuesta al crear un centro
type CreateCentroResponse struct {
	IDCentro int    `json:"id_centro"`
	IDMongo  int    `json:"id_mongo"`
	Message  string `json:"message"`
}

// CreateCentro crea un nuevo centro en MySQL y MongoDB
func CreateCentro(request CreateCentroRequest) (*CreateCentroResponse, error) {
	// 1. Crear el depot en MongoDB primero
	mongoID, err := createDepotInMongoDB(request.Latitud, request.Longitud, request.Neighborhood)
	if err != nil {
		return nil, fmt.Errorf("error creating depot in MongoDB: %v", err)
	}

	// 2. Crear el centro en MySQL con el ID de MongoDB
	centroID, err := createCentroInMySQL(request, mongoID)
	if err != nil {
		// Si falla MySQL, intentar eliminar el depot de MongoDB
		_ = deleteDepotFromMongoDB(mongoID)
		return nil, fmt.Errorf("error creating centro in MySQL: %v", err)
	}

	return &CreateCentroResponse{
		IDCentro: centroID,
		IDMongo:  mongoID,
		Message:  "Centro creado exitosamente",
	}, nil
}

// createCentroInMySQL crea un centro en MySQL y retorna su ID
func createCentroInMySQL(request CreateCentroRequest, mongoID int) (int, error) {
	if config.DB == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	query := `
		INSERT INTO Centro (id_tipo, id_mongo) 
		VALUES (?, ?)
	`

	result := config.DB.Exec(query, request.IdTipo, mongoID)
	if result.Error() != nil {
		return 0, fmt.Errorf("error inserting centro: %v", result.Error())
	}

	var centroID int64
	if result := config.DB.Raw("SELECT LAST_INSERT_ID()").Scan(&centroID); result.Error() != nil {
		return 0, fmt.Errorf("error getting inserted ID: %v", result.Error())
	}

	return int(centroID), nil
}

// DeleteCentro elimina un centro de MySQL y MongoDB
func DeleteCentro(centroID int, mongoID int) error {
	// Si se proporciona centroID, buscar el mongoID
	if centroID > 0 {
		var centro struct {
			IDMongo int `gorm:"column:id_mongo"`
		}
		if err := config.DB.Raw("SELECT id_mongo FROM Centro WHERE id_centro = ?", centroID).Scan(&centro).Error(); err != nil {
			return fmt.Errorf("error finding centro: %v", err)
		}
		if centro.IDMongo > 0 {
			mongoID = centro.IDMongo
		}
	}

	// 1. Eliminar de MySQL
	if err := deleteCentroFromMySQL(centroID, mongoID); err != nil {
		return fmt.Errorf("error deleting centro from MySQL: %v", err)
	}

	// 2. Eliminar de MongoDB
	if mongoID > 0 {
		if err := deleteDepotFromMongoDB(mongoID); err != nil {
			return fmt.Errorf("error deleting depot from MongoDB: %v", err)
		}
	}

	return nil
}

// deleteCentroFromMySQL elimina un centro de MySQL
func deleteCentroFromMySQL(centroID int, mongoID int) error {
	if config.DB == nil {
		return fmt.Errorf("database connection not available")
	}

	var query string
	var params []interface{}

	if centroID > 0 {
		query = "DELETE FROM Centro WHERE id_centro = ?"
		params = []interface{}{centroID}
	} else if mongoID > 0 {
		query = "DELETE FROM Centro WHERE id_mongo = ?"
		params = []interface{}{mongoID}
	} else {
		return fmt.Errorf("debe proporcionar centroID o mongoID para eliminar")
	}

	result := config.DB.Exec(query, params...)
	if result.Error() != nil {
		return fmt.Errorf("error deleting centro from MySQL: %v", result.Error())
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no se encontró el centro para eliminar")
	}

	return nil
}

// DEPRECATED: Las funciones de Neo4j ya no se usan para centros
// Los centros ahora se obtienen desde MongoDB (colección depot)
