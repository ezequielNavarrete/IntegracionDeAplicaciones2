package models

import "time"

type Tacho struct {
	IDTacho          int64      `gorm:"column:id_tacho;primaryKey"`
	IDTipo           int64      `gorm:"column:id_tipo"`
	IDEstado         int64      `gorm:"column:id_estado"`
	IDNeo            string     `gorm:"column:id_neo"`
	Capacidad        float64    `gorm:"column:capacidad"`
	Descripcion      *string    `gorm:"column:descripcion"`       // Opcional - para eventos culturales
	Observaciones    *string    `gorm:"column:observaciones"`     // Opcional - para eventos culturales
	FechaInstalacion *time.Time `gorm:"column:fecha_instalacion"` // Opcional - para eventos culturales
}

// TableName - nombre exacto de la tabla en MySQL
func (Tacho) TableName() string {
	return "Tacho"
}
