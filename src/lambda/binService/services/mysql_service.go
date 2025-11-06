package services

import (
	"fmt"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
)

// createTachoInMySQL crea un tacho en la tabla MySQL y retorna su ID
func createTachoInMySQL(request CreateTachoRequest, mongoID int) (int, error) {
	if config.DB == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	// Query para insertar en la tabla Tacho
	query := `
		INSERT INTO Tacho (id_tipo, id_estado, id_mongo, capacidad) 
		VALUES (?, ?, ?, ?)
	`

	result := config.DB.Exec(query,
		request.IdTipo,
		request.IdEstado,
		mongoID, // ID del documento de MongoDB
		request.Capacidad)

	if result.Error() != nil {
		return 0, fmt.Errorf("error inserting tacho: %v", result.Error())
	}

	// Obtener el ID generado por la inserción
	var tachoID int64
	if result := config.DB.Raw("SELECT LAST_INSERT_ID()").Scan(&tachoID); result.Error() != nil {
		return 0, fmt.Errorf("error getting inserted ID: %v", result.Error())
	}

	return int(tachoID), nil
}

// deleteTachoFromMySQL elimina un tacho de MySQL por ID o por ID de MongoDB
func deleteTachoFromMySQL(tachoID int, mongoID int) error {
	if config.DB == nil {
		return fmt.Errorf("database connection not available")
	}

	var query string
	var params []interface{}

	if tachoID > 0 {
		// Eliminar por ID del tacho
		query = "DELETE FROM Tacho WHERE id_tacho = ?"
		params = []interface{}{tachoID}
	} else if mongoID > 0 {
		// Eliminar por ID de MongoDB
		query = "DELETE FROM Tacho WHERE id_mongo = ?"
		params = []interface{}{mongoID}
	} else {
		return fmt.Errorf("debe proporcionar tachoID o mongoID para eliminar")
	}

	result := config.DB.Exec(query, params...)
	if result.Error() != nil {
		return fmt.Errorf("error deleting tacho from MySQL: %v", result.Error())
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no se encontró el tacho para eliminar")
	}

	return nil
}

/*

// getTachoByID obtiene un tacho de MySQL por su ID
func getTachoByID(tachoID int) (*TachoMySQL, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var tacho TachoMySQL
	result := config.DB.Raw("SELECT id_tacho, id_tipo, id_estado, id_neo, capacidad FROM Tacho WHERE id_tacho = ?", tachoID).Scan(&tacho)

	if result.Error() != nil {
		return nil, fmt.Errorf("error getting tacho: %v", result.Error())
	}

	// Verificamos si el registro está vacío en vez de usar RowsAffected()
	if tacho.ID == 0 {
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

	if result.Error() != nil {
		return nil, fmt.Errorf("error getting tacho: %v", result.Error())
	}

	if tacho.ID == 0 {
		return nil, fmt.Errorf("tacho not found")
	}

	return &tacho, nil
}*/

// TachoMySQL representa un tacho en la base de datos MySQL
type TachoMySQL struct {
	ID        int     `json:"id_tacho" gorm:"column:id_tacho"`
	IdTipo    int     `json:"id_tipo" gorm:"column:id_tipo"`
	IdEstado  int     `json:"id_estado" gorm:"column:id_estado"`
	IdMongo   int     `json:"id_mongo" gorm:"column:id_mongo"`
	Capacidad float64 `json:"capacidad" gorm:"column:capacidad"`
}

// CaracteristicaTacho representa una característica asociada a un tacho
type CaracteristicaTacho struct {
	Nombre    string `json:"nombre"`
	Estado    string `json:"estado"`
	Prioridad int    `json:"prioridad"`
}

// TachoCompleto representa un tacho con toda la información necesaria
type TachoCompleto struct {
	IDTacho         int                   `json:"id_tacho" gorm:"column:id_tacho"`
	Neighborhood    int                   `json:"neighborhood" gorm:"column:neighborhood"`
	Latitud         float64               `json:"latitud" gorm:"column:latitud"`
	Longitud        float64               `json:"longitud" gorm:"column:longitud"`
	Estado          string                `json:"estado" gorm:"column:estado"`
	Capacidad       float64               `json:"capacidad" gorm:"column:capacidad"`
	Caracteristicas []CaracteristicaTacho `json:"caracteristicas"`
}

// GetAllTachos obtiene todos los tachos con información completa
func GetAllTachos() ([]TachoCompleto, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Query optimizada que obtiene TODO en una sola consulta
	query := `
		SELECT 
			t.id_tacho,
			t.id_mongo,
			COALESCE(et.tipo_estado, 'activo') as estado,
			t.capacidad,
			ct.nombre as caracteristica_nombre,
			ec.nombre as caracteristica_estado,
			ec.prioridad as caracteristica_prioridad
		FROM Tacho t
		LEFT JOIN Estado_tacho et ON t.id_estado = et.id_estado
		LEFT JOIN Lista_caracteristica_tacho lct ON t.id_tacho = lct.id_tacho
		LEFT JOIN Caracteristica_tacho ct ON lct.id_caracteristica = ct.id_caracteristica
		LEFT JOIN Estado_caracteristica ec ON ct.id_estado_caracteristica = ec.id_estado_caracteristica
		ORDER BY t.id_tacho, ec.prioridad DESC
	`

	// Estructura temporal para leer los resultados con características
	type TachoCaracteristicaRow struct {
		IDTacho                 int     `gorm:"column:id_tacho"`
		IDMongo                 int     `gorm:"column:id_mongo"`
		Estado                  string  `gorm:"column:estado"`
		Capacidad               float64 `gorm:"column:capacidad"`
		CaracteristicaNombre    *string `gorm:"column:caracteristica_nombre"`
		CaracteristicaEstado    *string `gorm:"column:caracteristica_estado"`
		CaracteristicaPrioridad *int    `gorm:"column:caracteristica_prioridad"`
	}

	var rows []TachoCaracteristicaRow
	result := config.DB.Raw(query).Scan(&rows)
	if result.Error() != nil {
		return nil, fmt.Errorf("error getting tachos: %v", result.Error())
	}

	// Obtener coordenadas desde MongoDB una sola vez
	coordsMap, err := GetAllTachosFromMongoDB()
	if err != nil {
		return nil, fmt.Errorf("error getting coordinates from MongoDB: %v", err)
	}

	// Agrupar características por tacho
	tachosMap := make(map[int]*TachoCompleto)

	for _, row := range rows {
		// Si el tacho no existe en el map, crearlo
		if _, exists := tachosMap[row.IDTacho]; !exists {
			tacho := &TachoCompleto{
				IDTacho:         row.IDTacho,
				Estado:          row.Estado,
				Capacidad:       row.Capacidad,
				Neighborhood:    0,
				Latitud:         0,
				Longitud:        0,
				Caracteristicas: []CaracteristicaTacho{},
			}

			// Buscar coordenadas en MongoDB
			if mongoData, found := coordsMap[row.IDMongo]; found {
				tacho.Latitud = mongoData.Lat
				tacho.Longitud = mongoData.Lon
				tacho.Neighborhood = mongoData.Neighborhood
			}

			tachosMap[row.IDTacho] = tacho
		}

		// Agregar característica si existe
		if row.CaracteristicaNombre != nil && row.CaracteristicaEstado != nil && row.CaracteristicaPrioridad != nil {
			caracteristica := CaracteristicaTacho{
				Nombre:    *row.CaracteristicaNombre,
				Estado:    *row.CaracteristicaEstado,
				Prioridad: *row.CaracteristicaPrioridad,
			}
			tachosMap[row.IDTacho].Caracteristicas = append(tachosMap[row.IDTacho].Caracteristicas, caracteristica)
		}
	}

	// Convertir el map a slice
	tachos := make([]TachoCompleto, 0, len(tachosMap))
	for _, tacho := range tachosMap {
		tachos = append(tachos, *tacho)
	}

	return tachos, nil
}
