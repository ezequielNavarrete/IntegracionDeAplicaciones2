package services

import (
	"fmt"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
)

// createTachoInMySQL crea un tacho en la tabla MySQL y retorna su ID
func createTachoInMySQL(request CreateTachoRequest, customID string) (int, error) {
	if config.DB == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	// Query para insertar en la tabla Tacho
	query := `
		INSERT INTO Tacho (id_tipo, id_estado, id_neo, capacidad) 
		VALUES (?, ?, ?, ?)
	`

	result := config.DB.Exec(query,
		request.IdTipo,
		request.IdEstado,
		customID, // Usar el ID personalizado (direccion|barrio) en lugar del ID interno de Neo4j
		request.Capacidad)

	if result.Error != nil {
		return 0, fmt.Errorf("error inserting tacho: %v", result.Error)
	}

	// Obtener el ID generado por la inserción
	var tachoID int64
	if result := config.DB.Raw("SELECT LAST_INSERT_ID()").Scan(&tachoID); result.Error != nil {
		return 0, fmt.Errorf("error getting inserted ID: %v", result.Error)
	}

	return int(tachoID), nil
}

// deleteTachoFromMySQL elimina un tacho de MySQL por ID o por custom ID
func deleteTachoFromMySQL(tachoID int, customID string) error {
	if config.DB == nil {
		return fmt.Errorf("database connection not available")
	}

	var query string
	var params []interface{}

	if tachoID > 0 {
		// Eliminar por ID del tacho
		query = "DELETE FROM Tacho WHERE id_tacho = ?"
		params = []interface{}{tachoID}
	} else if customID != "" {
		// Eliminar por custom ID (que está guardado en id_neo)
		query = "DELETE FROM Tacho WHERE id_neo LIKE ?"
		params = []interface{}{"%" + customID + "%"}
	} else {
		return fmt.Errorf("debe proporcionar tachoID o customID para eliminar")
	}

	result := config.DB.Exec(query, params...)
	if result.Error != nil {
		return fmt.Errorf("error deleting tacho from MySQL: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no se encontró el tacho para eliminar")
	}

	return nil
}

// getTachoByID obtiene un tacho de MySQL por su ID
func getTachoByID(tachoID int) (*TachoMySQL, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var tacho TachoMySQL
	result := config.DB.Raw("SELECT id_tacho, id_tipo, id_estado, id_neo, capacidad FROM Tacho WHERE id_tacho = ?", tachoID).Scan(&tacho)

	if result.Error != nil {
		return nil, fmt.Errorf("error getting tacho: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("tacho not found")
	}

	return &tacho, nil
}

// getTachoByNeoID obtiene un tacho de MySQL por su ID de Neo4j
func getTachoByNeoID(neoNodeID string) (*TachoMySQL, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var tacho TachoMySQL
	result := config.DB.Raw("SELECT id_tacho, id_tipo, id_estado, id_neo, capacidad FROM Tacho WHERE id_neo = ?", neoNodeID).Scan(&tacho)

	if result.Error != nil {
		return nil, fmt.Errorf("error getting tacho: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("tacho not found")
	}

	return &tacho, nil
}

// TachoMySQL representa un tacho en la base de datos MySQL
type TachoMySQL struct {
	ID        int     `json:"id_tacho" gorm:"column:id_tacho"`
	IdTipo    int     `json:"id_tipo" gorm:"column:id_tipo"`
	IdEstado  int     `json:"id_estado" gorm:"column:id_estado"`
	IdNeo     string  `json:"id_neo" gorm:"column:id_neo"`
	Capacidad float64 `json:"capacidad" gorm:"column:capacidad"`
}

// TachoCompleto representa un tacho con toda la información necesaria
type TachoCompleto struct {
	IDTacho   int     `json:"id_tacho" gorm:"column:id_tacho"`
	Barrio    string  `json:"barrio" gorm:"column:barrio"`
	Direccion string  `json:"direccion" gorm:"column:direccion"`
	Latitud   float64 `json:"latitud" gorm:"column:latitud"`
	Longitud  float64 `json:"longitud" gorm:"column:longitud"`
	Estado    string  `json:"estado" gorm:"column:estado"`
	Capacidad float64 `json:"capacidad" gorm:"column:capacidad"`
}

// GetAllTachos obtiene todos los tachos con información completa
func GetAllTachos() ([]TachoCompleto, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var tachos []TachoCompleto

	// Query con JOIN para obtener el estado y datos de Neo4j parseados del id_neo
	query := `
		SELECT 
			t.id_tacho,
			SUBSTRING_INDEX(t.id_neo, '|', -1) as barrio,
			SUBSTRING_INDEX(t.id_neo, '|', 1) as direccion,
			t.id_neo as custom_id,
			et.tipo_estado as estado,
			t.capacidad
		FROM Tacho t
		INNER JOIN Estado_tacho et ON t.id_estado = et.id_estado_tacho
	`

	// Estructura temporal para obtener los datos de MySQL
	type TachoTemp struct {
		IDTacho   int     `gorm:"column:id_tacho"`
		Barrio    string  `gorm:"column:barrio"`
		Direccion string  `gorm:"column:direccion"`
		CustomID  string  `gorm:"column:custom_id"`
		Estado    string  `gorm:"column:estado"`
		Capacidad float64 `gorm:"column:capacidad"`
	}

	var tachosTemp []TachoTemp
	result := config.DB.Raw(query).Scan(&tachosTemp)
	if result.Error != nil {
		return nil, fmt.Errorf("error getting tachos: %v", result.Error)
	}

	// Obtener coordenadas de Neo4j
	coordsMap, err := GetAllTachosCoordinates()
	if err != nil {
		return nil, fmt.Errorf("error getting coordinates: %v", err)
	}

	// Combinar los datos
	for _, tachoTemp := range tachosTemp {
		tacho := TachoCompleto{
			IDTacho:   tachoTemp.IDTacho,
			Barrio:    tachoTemp.Barrio,
			Direccion: tachoTemp.Direccion,
			Estado:    tachoTemp.Estado,
			Capacidad: tachoTemp.Capacidad,
			Latitud:   0, // Default en caso de no encontrar
			Longitud:  0, // Default en caso de no encontrar
		}

		// Buscar las coordenadas en el map de Neo4j
		if coords, found := coordsMap[tachoTemp.CustomID]; found {
			tacho.Latitud = coords.Latitude
			tacho.Longitud = coords.Longitude
		}

		tachos = append(tachos, tacho)
	}

	return tachos, nil
}
