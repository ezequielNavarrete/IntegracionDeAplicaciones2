package services

import (
	"context"
	"fmt"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
)

// Estructura para representar un centro con información completa (MySQL + Neo4j)
type Centro struct {
	IDCentro   int    `json:"id_centro" gorm:"column:id_centro"`
	NombreTipo string `json:"nombre_tipo" gorm:"column:nombre_tipo"`
	IDNeo      string `json:"id_neo" gorm:"column:id_neo"`
	// Información adicional de Neo4j
	Nombre    string  `json:"nombre"`
	Barrio    string  `json:"barrio"`
	Direccion string  `json:"direccion"`
	Longitud  float64 `json:"longitud"`
	Latitud   float64 `json:"latitud"`
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

// GetAllCentros obtiene todos los centros con información de tipo (MySQL) y datos adicionales (Neo4j)
func GetAllCentros() (*CentrosResponse, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Estructura temporal para datos de MySQL
	type CentroMySQL struct {
		IDCentro   int    `gorm:"column:id_centro"`
		IDTipo     int    `gorm:"column:id_tipo"`
		NombreTipo string `gorm:"column:nombre_tipo"`
		IDNeo      string `gorm:"column:id_neo"`
	}

	var centrosMySQL []CentroMySQL

	// Query con JOIN para obtener información de MySQL
	query := `
		SELECT 
			c.id_centro,
			c.id_tipo,
			tc.nombre_tipo,
			c.id_neo
		FROM Centro c
		LEFT JOIN Tipo_centro tc ON c.id_tipo = tc.id_tipo
		ORDER BY c.id_centro ASC
	`

	if err := config.DB.Raw(query).Scan(&centrosMySQL).Error; err != nil {
		return nil, fmt.Errorf("error querying centros from MySQL: %v", err)
	}

	// Convertir a estructura completa y obtener datos de Neo4j
	var centros []Centro
	for _, centroMySQL := range centrosMySQL {
		centro := Centro{
			IDCentro:   centroMySQL.IDCentro,
			NombreTipo: centroMySQL.NombreTipo,
			IDNeo:      centroMySQL.IDNeo,
		}

		// Obtener información adicional de Neo4j
		if centroMySQL.IDNeo != "" {
			neoData, err := getCentroFromNeo4j(centroMySQL.IDNeo)
			if err != nil {
				// Log error pero continúa con otros centros
				fmt.Printf("Warning: Error getting Neo4j data for centro %s: %v\n", centroMySQL.IDNeo, err)
			} else {
				centro.Nombre = neoData.Nombre
				centro.Barrio = neoData.Barrio
				centro.Direccion = neoData.Direccion
				centro.Longitud = neoData.Longitud
				centro.Latitud = neoData.Latitud
			}
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
		IDNeo      string `gorm:"column:id_neo"`
	}

	var centroMySQL CentroMySQL

	// Query con JOIN para obtener información de MySQL
	query := `
		SELECT 
			c.id_centro,
			c.id_tipo,
			tc.nombre_tipo,
			c.id_neo
		FROM Centro c
		LEFT JOIN Tipo_centro tc ON c.id_tipo = tc.id_tipo
		WHERE c.id_centro = ?
	`

	if err := config.DB.Raw(query, centroID).Scan(&centroMySQL).Error; err != nil {
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
		IDNeo:      centroMySQL.IDNeo,
	}

	// Obtener información adicional de Neo4j si existe IDNeo
	if centroMySQL.IDNeo != "" {
		neoData, err := getCentroFromNeo4j(centroMySQL.IDNeo)
		if err != nil {
			return nil, fmt.Errorf("error getting Neo4j data for centro %s: %v", centroMySQL.IDNeo, err)
		}
		centro.Nombre = neoData.Nombre
		centro.Barrio = neoData.Barrio
		centro.Direccion = neoData.Direccion
		centro.Longitud = neoData.Longitud
		centro.Latitud = neoData.Latitud
	}

	return &CentroResponse{
		Centro: centro,
	}, nil
}

// Estructura para datos de Neo4j
type CentroNeo4jData struct {
	Nombre    string  `json:"nombre"`
	Barrio    string  `json:"barrio"`
	Direccion string  `json:"direccion"`
	Longitud  float64 `json:"longitud"`
	Latitud   float64 `json:"latitud"`
}

// getCentroFromNeo4j obtiene información adicional del centro desde Neo4j
func getCentroFromNeo4j(idNeo string) (*CentroNeo4jData, error) {
	session, err := config.GetNeo4jSession()
	if err != nil {
		return nil, fmt.Errorf("error getting Neo4j session: %v", err)
	}
	defer session.Close(context.Background())

	// Query para obtener información del centro en Neo4j incluyendo coordenadas del campo location
	query := `
		MATCH (c) 
		WHERE c.id = $id_neo
		RETURN c.nombre as nombre, c.barrio as barrio, c.direccion as direccion, 
		       c.location.longitude as longitud, c.location.latitude as latitud
	`

	result, err := session.Run(context.Background(), query, map[string]interface{}{
		"id_neo": idNeo,
	})
	if err != nil {
		return nil, fmt.Errorf("error executing Neo4j query: %v", err)
	}

	if result.Next(context.Background()) {
		record := result.Record()

		// Obtener valores con verificación de nulos
		nombre, _ := record.Get("nombre")
		barrio, _ := record.Get("barrio")
		direccion, _ := record.Get("direccion")
		longitud, _ := record.Get("longitud")
		latitud, _ := record.Get("latitud")

		return &CentroNeo4jData{
			Nombre:    getStringValue(nombre),
			Barrio:    getStringValue(barrio),
			Direccion: getStringValue(direccion),
			Longitud:  getFloatValue(longitud),
			Latitud:   getFloatValue(latitud),
		}, nil
	}

	return nil, fmt.Errorf("centro not found in Neo4j with id: %s", idNeo)
}

// Funciones auxiliares para manejo seguro de valores
func getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", value)
}

func getFloatValue(value interface{}) float64 {
	if value == nil {
		return 0.0
	}
	if f, ok := value.(float64); ok {
		return f
	}
	if i, ok := value.(int64); ok {
		return float64(i)
	}
	return 0.0
}
