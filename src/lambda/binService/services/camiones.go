package services

import (
	"fmt"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
)

// Estructura para representar un camión con información completa
type Camion struct {
	IDCamion   int    `json:"id_camion" gorm:"column:id_camion"`
	Capacidad  int    `json:"capacidad" gorm:"column:capacidad"`
	Modelo     string `json:"modelo" gorm:"column:modelo"`
	Marca      string `json:"marca" gorm:"column:marca"`
	Matricula  int    `json:"matricula" gorm:"column:matricula"`
	NombreTipo string `json:"nombre_tipo" gorm:"column:nombre_tipo"`
	TipoEstado string `json:"tipo_estado" gorm:"column:tipo_estado"`
}

// Estructura para respuesta de camiones
type CamionesResponse struct {
	Camiones []Camion `json:"camiones"`
	Total    int      `json:"total"`
}

// Estructura para respuesta de un camión individual
type CamionResponse struct {
	Camion Camion `json:"camion"`
}

// GetAllCamiones obtiene todos los camiones con información de tipo y estado
func GetAllCamiones() (*CamionesResponse, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var camiones []Camion

	// Query con JOINs para obtener información completa
	query := `
		SELECT 
			c.id_camion,
			c.capacidad,
			c.modelo,
			c.marca,
			c.matricula,
			c.id_tipo,
			tc.nombre_tipo,
			c.id_estado,
			ec.tipo_estado
		FROM Camiones c
		LEFT JOIN Tipo_camion tc ON c.id_tipo = tc.id_tipo
		LEFT JOIN Estado_camion ec ON c.id_estado = ec.id_estado
		ORDER BY c.id_camion ASC
	`

	result := config.DB.Raw(query).Scan(&camiones)
	if result.Error() != nil {
		return nil, fmt.Errorf("error querying camiones: %v", result.Error())
	}

	return &CamionesResponse{
		Camiones: camiones,
		Total:    len(camiones),
	}, nil
}

// GetCamionByID obtiene un camión específico por ID con información de tipo y estado
func GetCamionByID(camionID int) (*CamionResponse, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}
	var camion Camion

	// Query con JOINs para obtener información completa de un camión específico
	query := `
		SELECT 
			c.id_camion,
			c.capacidad,
			c.modelo,
			c.marca,
			c.matricula,
			c.id_tipo,
			tc.nombre_tipo,
			c.id_estado,
			ec.tipo_estado
		FROM Camiones c
		LEFT JOIN Tipo_camion tc ON c.id_tipo = tc.id_tipo
		LEFT JOIN Estado_camion ec ON c.id_estado = ec.id_estado
		WHERE c.id_camion = ?
	`
	result := config.DB.Raw(query, camionID).Scan(&camion)
	if result.Error() != nil {
		return nil, fmt.Errorf("error querying camion: %v", result.Error())
	}

	// Verificar si se encontró el camión
	if camion.IDCamion == 0 {
		return nil, fmt.Errorf("camion with ID %d not found", camionID)
	}

	return &CamionResponse{Camion: camion}, nil
}

// =====================
// Sección MySQL (creación/eliminación definitivas)
// =====================

// CreateCamionMySQLRequest define los campos necesarios para crear un camión en MySQL
type CreateCamionMySQLRequest struct {
	Capacidad int    `json:"capacidad" binding:"required,min=1"`
	Modelo    string `json:"modelo" binding:"required,min=2"`
	Marca     string `json:"marca" binding:"required,min=2"`
	Matricula int    `json:"matricula" binding:"required,min=1"`
	IdTipo    int    `json:"id_tipo" binding:"required,min=1" enums:"1,2,3,4"`   // 1 Basura, 2 Reciclaje, 3 Limpieza, 4 Especial
	IdEstado  int    `json:"id_estado" binding:"required,min=1" enums:"1,2,3,4"` // 1 Operativo, 2 En uso, 3 En mantenimiento, 4 Fuera de servicio
}

// CreateCamionMySQL inserta el camión y devuelve el registro enriquecido con JOINs
func CreateCamionMySQL(req CreateCamionMySQLRequest) (*Camion, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Insertar en Camiones
	insert := `
		INSERT INTO Camiones (capacidad, modelo, marca, matricula, id_tipo, id_estado)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	res := config.DB.Exec(insert, req.Capacidad, req.Modelo, req.Marca, req.Matricula, req.IdTipo, req.IdEstado)
	if res.Error() != nil {
		return nil, fmt.Errorf("error inserting camion: %v", res.Error())
	}

	// Obtener ID generado
	var newID int
	if r := config.DB.Raw("SELECT LAST_INSERT_ID()").Scan(&newID); r.Error() != nil {
		return nil, fmt.Errorf("error getting inserted ID: %v", r.Error())
	}

	// Devolver con JOINs el registro completo
	var camion Camion
	query := `
		SELECT 
			c.id_camion,
			c.capacidad,
			c.modelo,
			c.marca,
			c.matricula,
			c.id_tipo,
			tc.nombre_tipo,
			c.id_estado,
			ec.tipo_estado
		FROM Camiones c
		LEFT JOIN Tipo_camion tc ON c.id_tipo = tc.id_tipo
		LEFT JOIN Estado_camion ec ON c.id_estado = ec.id_estado
		WHERE c.id_camion = ?
	`
	if r := config.DB.Raw(query, newID).Scan(&camion); r.Error() != nil {
		return nil, fmt.Errorf("error querying created camion: %v", r.Error())
	}

	return &camion, nil
}

// DeleteCamionMySQL elimina un camión por ID
func DeleteCamionMySQL(id int) error {
	if config.DB == nil {
		return fmt.Errorf("database connection not available")
	}
	if id <= 0 {
		return fmt.Errorf("id inválido")
	}

	del := "DELETE FROM Camiones WHERE id_camion = ?"
	res := config.DB.Exec(del, id)
	if res.Error() != nil {
		return fmt.Errorf("error deleting camion: %v", res.Error())
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("camion %d not found", id)
	}
	return nil
}
