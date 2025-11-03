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
	Matricula  string `json:"matricula" gorm:"column:matricula"`
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

	return &CamionResponse{
		Camion: camion,
	}, nil
}
