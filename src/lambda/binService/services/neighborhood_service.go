package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

// GetNeighborhoodByCoordinates consulta el servicio externo para obtener la comuna (neighborhood)
// a partir de lat/lng. Devuelve un *int (nil si no se pudo determinar).
func GetNeighborhoodByCoordinates(lat, lon float64) (*int, error) {
	if lat == 0 && lon == 0 { // Coordenadas inválidas
		return nil, errors.New("coordenadas vacías")
	}

	url := os.Getenv("NEIGHBORHOOD_SERVICE_URL")
	if url == "" {
		return nil, errors.New("NEIGHBORHOOD_SERVICE_URL no configurado")
	}

	// Payload requerido por el servicio externo
	reqBody := map[string]interface{}{
		"mode": "coordinates",
		"point": map[string]float64{
			"lat": lat,
			"lon": lon,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error serializando request neighborhood: %w", err)
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error llamando neighborhood service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("neighborhood service status %d", resp.StatusCode)
	}

	// Intentar decodificar respuesta flexible
	var generic map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&generic); err != nil {
		return nil, fmt.Errorf("error decodificando respuesta neighborhood: %w", err)
	}

	// Buscar campo neighborhood (o comuna) como número
	if v, ok := generic["neighborhood"]; ok {
		switch val := v.(type) {
		case float64:
			n := int(val)
			return &n, nil
		case int:
			n := val
			return &n, nil
		}
	}
	if v, ok := generic["comuna"]; ok { // fallback alternativa
		switch val := v.(type) {
		case float64:
			n := int(val)
			return &n, nil
		case int:
			n := val
			return &n, nil
		}
	}

	// Si no encontramos el campo esperado, devolver nil sin error grave
	return nil, fmt.Errorf("respuesta sin campo neighborhood/comuna")
}
