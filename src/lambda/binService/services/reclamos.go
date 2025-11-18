package services

import (
	"fmt"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
)

// Reclamo representa un reclamo en la base de datos
type Reclamo struct {
	IDReclamo        int       `json:"id_reclamo" gorm:"column:id_reclamo;primaryKey"`
	IDPersona        int       `json:"id_persona" gorm:"column:id_persona"`
	IDSubcategoria   *int      `json:"id_subcategoria" gorm:"column:id_subcategoria"`
	Titulo           string    `json:"titulo" gorm:"column:titulo"`
	Descripcion      string    `json:"descripcion" gorm:"column:descripcion"`
	Prioridad        string    `json:"prioridad" gorm:"column:prioridad"`
	Estado           string    `json:"estado" gorm:"column:estado"`
	Direccion        string    `json:"direccion" gorm:"column:direccion"`
	Lat              float64   `json:"lat" gorm:"column:lat"`
	Lng              float64   `json:"lng" gorm:"column:lng"`
	Fecha            time.Time `json:"fecha" gorm:"column:fecha"`
	IDReclamoExterno *int      `json:"id_reclamo_externo,omitempty" gorm:"column:id_reclamo_externo"` // ID del reclamo en el sistema externo
}

// CreateReclamoRequest representa la solicitud para crear un reclamo
type CreateReclamoRequest struct {
	IDPersona      int     `json:"id_persona" binding:"required"`
	IDSubcategoria *int    `json:"id_subcategoria"`
	Titulo         string  `json:"titulo" binding:"required"`
	Descripcion    string  `json:"descripcion" binding:"required"`
	Prioridad      string  `json:"prioridad"`
	Estado         string  `json:"estado"`
	Direccion      string  `json:"direccion" binding:"required"`
	Lat            float64 `json:"lat" binding:"required"`
	Lng            float64 `json:"lng" binding:"required"`
}

// CreateReclamoResponse representa la respuesta al crear un reclamo
type CreateReclamoResponse struct {
	IDReclamo int       `json:"id_reclamo"`
	Fecha     time.Time `json:"fecha"`
	Message   string    `json:"message"`
}

// GetAllReclamos obtiene todos los reclamos
func GetAllReclamos() ([]Reclamo, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var reclamos []Reclamo
	query := "SELECT * FROM Reclamos ORDER BY fecha DESC"
	result := config.DB.Raw(query).Scan(&reclamos)
	if result.Error() != nil {
		return nil, fmt.Errorf("error getting reclamos: %v", result.Error())
	}

	return reclamos, nil
}

// GetReclamoByID obtiene un reclamo específico por ID
func GetReclamoByID(reclamoID int) (*Reclamo, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var reclamo Reclamo
	query := "SELECT * FROM Reclamos WHERE id_reclamo = ?"
	result := config.DB.Raw(query, reclamoID).Scan(&reclamo)
	if result.Error() != nil {
		return nil, fmt.Errorf("error getting reclamo: %v", result.Error())
	}

	if reclamo.IDReclamo == 0 {
		return nil, fmt.Errorf("reclamo not found")
	}

	return &reclamo, nil
}

// CreateReclamo crea un nuevo reclamo
func CreateReclamo(request CreateReclamoRequest) (*CreateReclamoResponse, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Crear el reclamo con la fecha actual
	fecha := time.Now()

	// Valores por defecto
	prioridad := request.Prioridad
	if prioridad == "" {
		prioridad = "Media"
	}

	estado := request.Estado
	if estado == "" {
		estado = "Pendiente"
	}

	query := `
		INSERT INTO Reclamos (id_persona, id_subcategoria, titulo, descripcion, prioridad, estado, direccion, lat, lng, fecha)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result := config.DB.Exec(query,
		request.IDPersona,
		request.IDSubcategoria,
		request.Titulo,
		request.Descripcion,
		prioridad,
		estado,
		request.Direccion,
		request.Lat,
		request.Lng,
		fecha,
	)

	if result.Error() != nil {
		return nil, fmt.Errorf("error creating reclamo: %v", result.Error())
	}

	// Obtener el ID generado
	var reclamoID int64
	if result := config.DB.Raw("SELECT LAST_INSERT_ID()").Scan(&reclamoID); result.Error() != nil {
		return nil, fmt.Errorf("error getting inserted ID: %v", result.Error())
	}

	return &CreateReclamoResponse{
		IDReclamo: int(reclamoID),
		Fecha:     fecha,
		Message:   "Reclamo creado exitosamente",
	}, nil
}

// DeleteReclamo elimina un reclamo por ID
func DeleteReclamo(reclamoID int) error {
	if config.DB == nil {
		return fmt.Errorf("database connection not available")
	}

	result := config.DB.Exec("DELETE FROM Reclamos WHERE id_reclamo = ?", reclamoID)
	if result.Error() != nil {
		return fmt.Errorf("error deleting reclamo: %v", result.Error())
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("reclamo not found")
	}

	return nil
}

// UpdateReclamoEstado actualiza el estado de un reclamo
func UpdateReclamoEstado(reclamoID int, estado string) error {
	if config.DB == nil {
		return fmt.Errorf("database connection not available")
	}

	result := config.DB.Exec("UPDATE Reclamos SET estado = ? WHERE id_reclamo = ?", estado, reclamoID)
	if result.Error() != nil {
		return fmt.Errorf("error updating reclamo estado: %v", result.Error())
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("reclamo not found")
	}

	return nil
}
