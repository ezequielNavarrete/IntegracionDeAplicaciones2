package services

import (
	"context"
	"fmt"
	"strings"
	"time"

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

// =====================
// Sección Redis (creación temporal de camiones antes de migrar a MySQL)
// =====================

// CreateCamionRequestRedis define los campos necesarios para crear un camión en Redis.
// Se usan los nombres directos de tipo y estado (no IDs) para simplificar la futura migración.
type CreateCamionRequestRedis struct {
	Capacidad  int    `json:"capacidad" binding:"required,min=1"`
	Modelo     string `json:"modelo" binding:"required,min=2"`
	Marca      string `json:"marca" binding:"required,min=2"`
	Matricula  string `json:"matricula" binding:"required,min=1"`
	NombreTipo string `json:"nombre_tipo" binding:"required"`  // Basura | Reciclaje | Limpieza | Especial
	TipoEstado string `json:"tipo_estado" binding:"required"`  // Operativo | En uso | En mantenimiento | Fuera de servicio
}

var allowedTipos = []string{"Basura", "Reciclaje", "Limpieza", "Especial"}
var allowedEstados = []string{"Operativo", "En uso", "En mantenimiento", "Fuera de servicio"}

func isAllowed(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

// CreateCamionRedis crea y almacena un camión en Redis.
// Patrón de almacenamiento:
//   - Secuencia: camion:id_seq (INCR)
//   - Hash por camión: camion:<id>
//   - Lista de IDs: camiones:ids (RPUSH para mantener orden de inserción)
// Nota: No se modifica la obtención actual (que lee de MySQL). Los camiones creados aquí
//       todavía no aparecen en GetAllCamiones hasta la migración completa.
func CreateCamionRedis(req CreateCamionRequestRedis) (*Camion, error) {
	if config.RedisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	// Validaciones de dominio sencillas (sin usar IDs numéricos).
	if !isAllowed(req.NombreTipo, allowedTipos) {
		return nil, fmt.Errorf("nombre_tipo inválido. Valores permitidos: %s", strings.Join(allowedTipos, ", "))
	}
	if !isAllowed(req.TipoEstado, allowedEstados) {
		return nil, fmt.Errorf("tipo_estado inválido. Valores permitidos: %s", strings.Join(allowedEstados, ", "))
	}

	// Generar nuevo ID secuencial.
	id, err := config.RedisClient.Incr(context.Background(), "camion:id_seq").Result()
	if err != nil {
		return nil, fmt.Errorf("error incrementando secuencia: %v", err)
	}

	key := fmt.Sprintf("camion:%d", id)
	camion := &Camion{
		IDCamion:   int(id),
		Capacidad:  req.Capacidad,
		Modelo:     req.Modelo,
		Marca:      req.Marca,
		Matricula:  req.Matricula,
		NombreTipo: req.NombreTipo,
		TipoEstado: req.TipoEstado,
	}

	// Hash de almacenamiento.
	data := map[string]interface{}{
		"id_camion":   camion.IDCamion,
		"capacidad":   camion.Capacidad,
		"modelo":      camion.Modelo,
		"marca":       camion.Marca,
		"matricula":   camion.Matricula,
		"nombre_tipo": camion.NombreTipo,
		"tipo_estado": camion.TipoEstado,
		"created_at":  time.Now().UTC().Format(time.RFC3339), // campo auxiliar, no expuesto en struct
	}

	if err := config.RedisClient.HSet(context.Background(), key, data).Err(); err != nil {
		return nil, fmt.Errorf("error guardando hash de camion: %v", err)
	}

	// Agregar ID a la lista de IDs para futura iteración.
	if err := config.RedisClient.RPush(context.Background(), "camiones:ids", camion.IDCamion).Err(); err != nil {
		return nil, fmt.Errorf("error agregando ID a lista de camiones: %v", err)
	}

	return camion, nil
}

// DeleteCamionRedis elimina un camión de Redis (hash + referencia en lista camiones:ids)
func DeleteCamionRedis(id int) error {
	if config.RedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	if id <= 0 {
		return fmt.Errorf("id inválido")
	}

	key := fmt.Sprintf("camion:%d", id)
	exists, err := config.RedisClient.Exists(context.Background(), key).Result()
	if err != nil {
		return fmt.Errorf("error verificando existencia: %v", err)
	}
	if exists == 0 {
		return fmt.Errorf("camion %d not found", id)
	}

	// Borrar hash
	if err := config.RedisClient.Del(context.Background(), key).Err(); err != nil {
		return fmt.Errorf("error eliminando hash: %v", err)
	}

	// Remover ID de la lista (si estaba). No es crítico si no se encuentra.
	if err := config.RedisClient.LRem(context.Background(), "camiones:ids", 0, id).Err(); err != nil {
		// Log de error silencioso: se mantiene éxito de eliminación de hash.
		return fmt.Errorf("hash eliminado pero error removiendo de lista: %v", err)
	}

	return nil
}
