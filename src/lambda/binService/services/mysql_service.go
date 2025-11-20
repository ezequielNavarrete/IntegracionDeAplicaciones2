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

	// Asignar características por defecto con estado "Nulo" (prioridad 0)
	if err := assignDefaultCaracteristicasToTacho(int(tachoID)); err != nil {
		// Si falla la asignación, eliminar el tacho creado para mantener consistencia
		_ = deleteTachoFromMySQL(int(tachoID), 0)
		return 0, fmt.Errorf("error assigning default caracteristicas: %v", err)
	}

	return int(tachoID), nil
}

// assignDefaultCaracteristicasToTacho asigna todas las características con estado "Nulo" a un tacho nuevo
func assignDefaultCaracteristicasToTacho(tachoID int) error {
	if config.DB == nil {
		return fmt.Errorf("database connection not available")
	}

	// Obtener todas las características y su estado "Nulo" (prioridad 0)
	query := `
		INSERT INTO Lista_caracteristica_tacho (id_tacho, id_caracteristica)
		SELECT ?, ct.id_caracteristica
		FROM Caracteristica_tacho ct
		INNER JOIN Estado_caracteristica ec ON ct.id_estado_caracteristica = ec.id_estado_caracteristica
		WHERE ec.prioridad = 0
	`

	result := config.DB.Exec(query, tachoID)
	if result.Error() != nil {
		return fmt.Errorf("error inserting default caracteristicas: %v", result.Error())
	}

	return nil
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
	IDTipo          int                   `json:"id_tipo" gorm:"column:id_tipo"`
	TipoTacho       string                `json:"tipo_tacho" gorm:"column:tipo_tacho"`
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
			t.id_tipo,
			COALESCE(tt.nombre_tipo, 'Desconocido') as tipo_tacho,
			t.id_mongo,
			COALESCE(et.tipo_estado, 'activo') as estado,
			t.capacidad,
			ct.nombre as caracteristica_nombre,
			ec.nombre as caracteristica_estado,
			ec.prioridad as caracteristica_prioridad
		FROM Tacho t
		LEFT JOIN Tipo_tacho tt ON t.id_tipo = tt.id_tipo
		LEFT JOIN Estado_tacho et ON t.id_estado = et.id_estado
		LEFT JOIN Lista_caracteristica_tacho lct ON t.id_tacho = lct.id_tacho
		LEFT JOIN Caracteristica_tacho ct ON lct.id_caracteristica = ct.id_caracteristica
		LEFT JOIN Estado_caracteristica ec ON ct.id_estado_caracteristica = ec.id_estado_caracteristica
		ORDER BY t.id_tacho, ec.prioridad DESC
	`

	// Estructura temporal para leer los resultados con características
	type TachoCaracteristicaRow struct {
		IDTacho                 int     `gorm:"column:id_tacho"`
		IDTipo                  int     `gorm:"column:id_tipo"`
		TipoTacho               string  `gorm:"column:tipo_tacho"`
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
				IDTipo:          row.IDTipo,
				TipoTacho:       row.TipoTacho,
				Estado:          row.Estado,
				Capacidad:       row.Capacidad,
				Neighborhood:    0,
				Latitud:         0,
				Longitud:        0,
				Caracteristicas: []CaracteristicaTacho{},
			}

		// Buscar coordenadas en MongoDB
		if mongoData, found := coordsMap[row.IDMongo]; found {
			// MongoDB tiene coordenadas invertidas, las intercambiamos antes de devolver
			latReal, lonReal := SwapCoordinates(mongoData.Lat, mongoData.Lon)
			tacho.Latitud = latReal
			tacho.Longitud = lonReal
			tacho.Neighborhood = mongoData.Neighborhood
		}

		tachosMap[row.IDTacho] = tacho
	}		// Agregar característica si existe
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

// GetTachoByID obtiene un tacho específico por ID con todas sus características
func GetTachoByID(tachoID int) (*TachoCompleto, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Query para obtener el tacho con sus características
	query := `
		SELECT 
			t.id_tacho,
			t.id_tipo,
			COALESCE(tt.nombre_tipo, 'Desconocido') as tipo_tacho,
			t.id_mongo,
			COALESCE(et.tipo_estado, 'activo') as estado,
			t.capacidad,
			ct.nombre as caracteristica_nombre,
			ec.nombre as caracteristica_estado,
			ec.prioridad as caracteristica_prioridad
		FROM Tacho t
		LEFT JOIN Tipo_tacho tt ON t.id_tipo = tt.id_tipo
		LEFT JOIN Estado_tacho et ON t.id_estado = et.id_estado
		LEFT JOIN Lista_caracteristica_tacho lct ON t.id_tacho = lct.id_tacho
		LEFT JOIN Caracteristica_tacho ct ON lct.id_caracteristica = ct.id_caracteristica
		LEFT JOIN Estado_caracteristica ec ON ct.id_estado_caracteristica = ec.id_estado_caracteristica
		WHERE t.id_tacho = ?
		ORDER BY ec.prioridad DESC
	`

	// Estructura temporal para leer los resultados con características
	type TachoCaracteristicaRow struct {
		IDTacho                 int     `gorm:"column:id_tacho"`
		IDTipo                  int     `gorm:"column:id_tipo"`
		TipoTacho               string  `gorm:"column:tipo_tacho"`
		IDMongo                 int     `gorm:"column:id_mongo"`
		Estado                  string  `gorm:"column:estado"`
		Capacidad               float64 `gorm:"column:capacidad"`
		CaracteristicaNombre    *string `gorm:"column:caracteristica_nombre"`
		CaracteristicaEstado    *string `gorm:"column:caracteristica_estado"`
		CaracteristicaPrioridad *int    `gorm:"column:caracteristica_prioridad"`
	}

	var rows []TachoCaracteristicaRow
	result := config.DB.Raw(query, tachoID).Scan(&rows)
	if result.Error() != nil {
		return nil, fmt.Errorf("error getting tacho: %v", result.Error())
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("tacho with ID %d not found", tachoID)
	}

	// Obtener el primer row para datos básicos
	firstRow := rows[0]

	// Crear el tacho con datos básicos
	tacho := &TachoCompleto{
		IDTacho:         firstRow.IDTacho,
		IDTipo:          firstRow.IDTipo,
		TipoTacho:       firstRow.TipoTacho,
		Estado:          firstRow.Estado,
		Capacidad:       firstRow.Capacidad,
		Neighborhood:    0,
		Latitud:         0,
		Longitud:        0,
		Caracteristicas: []CaracteristicaTacho{},
	}

	// Obtener coordenadas desde MongoDB
	if firstRow.IDMongo > 0 {
		coordsMap, err := GetAllTachosFromMongoDB()
		if err == nil {
			if mongoData, found := coordsMap[firstRow.IDMongo]; found {
				// MongoDB tiene coordenadas invertidas, las intercambiamos antes de devolver
				latReal, lonReal := SwapCoordinates(mongoData.Lat, mongoData.Lon)
				tacho.Latitud = latReal
				tacho.Longitud = lonReal
				tacho.Neighborhood = mongoData.Neighborhood
			}
		}
	}

	// Agregar todas las características
	for _, row := range rows {
		if row.CaracteristicaNombre != nil && row.CaracteristicaEstado != nil && row.CaracteristicaPrioridad != nil {
			caracteristica := CaracteristicaTacho{
				Nombre:    *row.CaracteristicaNombre,
				Estado:    *row.CaracteristicaEstado,
				Prioridad: *row.CaracteristicaPrioridad,
			}
			tacho.Caracteristicas = append(tacho.Caracteristicas, caracteristica)
		}
	}

	return tacho, nil
}

// UpdateCaracteristicaRequest representa una característica a actualizar
type UpdateCaracteristicaRequest struct {
	Nombre    string `json:"nombre"`    // "Humedad", "Olor", "Llenado", "Tipo de residuo", "Temperatura"
	Prioridad int    `json:"prioridad"` // 0-4: Nulo, Bajo, Medio, Alto, Urgente
}

// UpdateTachoCaracteristicasRequest representa la solicitud para actualizar características
type UpdateTachoCaracteristicasRequest struct {
	Caracteristicas []UpdateCaracteristicaRequest `json:"caracteristicas"`
}

// UpdateTachoCaracteristicas actualiza las características de un tacho
func UpdateTachoCaracteristicas(tachoID int, request UpdateTachoCaracteristicasRequest) error {
	if config.DB == nil {
		return fmt.Errorf("database connection not available")
	}

	// Verificar que el tacho existe
	var exists int
	result := config.DB.Raw("SELECT COUNT(*) FROM Tacho WHERE id_tacho = ?", tachoID).Scan(&exists)
	if result.Error() != nil || exists == 0 {
		return fmt.Errorf("tacho not found")
	}

	// Actualizar cada característica
	for _, carac := range request.Caracteristicas {
		// Validar prioridad (0-4)
		if carac.Prioridad < 0 || carac.Prioridad > 4 {
			return fmt.Errorf("prioridad debe estar entre 0 y 4 para característica '%s'", carac.Nombre)
		}

		// Buscar el id_caracteristica por nombre
		var idCaracteristica int
		queryCarac := `
			SELECT ct.id_caracteristica 
			FROM Caracteristica_tacho ct 
			WHERE ct.nombre = ?
			LIMIT 1
		`
		result := config.DB.Raw(queryCarac, carac.Nombre).Scan(&idCaracteristica)
		if result.Error() != nil || idCaracteristica == 0 {
			return fmt.Errorf("característica '%s' no encontrada", carac.Nombre)
		}

		// Buscar el id_estado_caracteristica con la prioridad especificada para esta característica
		var idEstadoCaracteristica int
		queryEstado := `
			SELECT ec.id_estado_caracteristica
			FROM Estado_caracteristica ec
			WHERE ec.prioridad = ?
			LIMIT 1
		`
		result = config.DB.Raw(queryEstado, carac.Prioridad).Scan(&idEstadoCaracteristica)
		if result.Error() != nil || idEstadoCaracteristica == 0 {
			return fmt.Errorf("estado con prioridad %d no encontrado", carac.Prioridad)
		}

		// Actualizar Caracteristica_tacho con el nuevo estado
		updateQuery := `
			UPDATE Caracteristica_tacho ct
			INNER JOIN Lista_caracteristica_tacho lct ON ct.id_caracteristica = lct.id_caracteristica
			SET ct.id_estado_caracteristica = ?
			WHERE lct.id_tacho = ? AND ct.nombre = ?
		`
		updateResult := config.DB.Exec(updateQuery, idEstadoCaracteristica, tachoID, carac.Nombre)
		if updateResult.Error() != nil {
			return fmt.Errorf("error actualizando característica '%s': %v", carac.Nombre, updateResult.Error())
		}
	}

	return nil
}
