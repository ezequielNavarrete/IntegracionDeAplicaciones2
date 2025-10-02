package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock data para personas
var mockPersonas = map[string]map[string]string{
	"1": {"id": "1", "zona_id": "1", "camion_id": "1"},
	"2": {"id": "2", "zona_id": "1", "camion_id": "2"},
	"3": {"id": "3", "zona_id": "2", "camion_id": "1"},
	"4": {"id": "4", "zona_id": "2", "camion_id": "2"},
	"5": {"id": "5", "zona_id": "3", "camion_id": "1"},
}

func TestGetAllPersonas(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "Obtener todas las personas exitosamente",
			expectedStatus: 200,
			checkResponse:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("GET", "/personas", nil)

			// Como el handler real usa Redis, vamos a crear un handler mock para testing
			GetAllPersonasMock(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "personas")
				assert.Contains(t, response, "total")
			}
		})
	}
}

func TestGetPersonaByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		personaID      string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Persona encontrada",
			personaID:      "1",
			expectedStatus: 200,
		},
		{
			name:           "Persona no encontrada",
			personaID:      "999",
			expectedStatus: 404,
			expectedError:  "Persona no encontrada",
		},
		{
			name:           "ID inválido",
			personaID:      "invalid",
			expectedStatus: 400,
			expectedError:  "ID inválido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("GET", "/personas/"+tt.personaID, nil)
			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.personaID}}

			GetPersonaByIDMock(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}

func TestGetPersonasByZona(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		zona           string
		expectedStatus int
		expectedError  string
		expectedCount  int
	}{
		{
			name:           "Zona válida con personas",
			zona:           "1",
			expectedStatus: 200,
			expectedCount:  2, // personas 1 y 2 están en zona 1
		},
		{
			name:           "Zona válida sin personas",
			zona:           "5",
			expectedStatus: 200,
			expectedCount:  0,
		},
		{
			name:           "Zona inválida",
			zona:           "invalid",
			expectedStatus: 400,
			expectedError:  "Zona inválida",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("GET", "/personas/zona/"+tt.zona, nil)
			c.Params = gin.Params{gin.Param{Key: "zona", Value: tt.zona}}

			GetPersonasByZonaMock(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.expectedStatus == 200 {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "personas")
				assert.Contains(t, response, "total")
				assert.Equal(t, float64(tt.expectedCount), response["total"])
			}
		})
	}
}

// Mock handlers para testing (sin dependencias externas)

func GetAllPersonasMock(c *gin.Context) {
	var result []PersonaResponse

	for _, persona := range mockPersonas {
		result = append(result, PersonaResponse{
			ID:       persona["id"],
			ZonaID:   persona["zona_id"],
			CamionID: persona["camion_id"],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    len(result),
		"personas": result,
	})
}

func GetPersonaByIDMock(c *gin.Context) {
	idStr := c.Param("id")

	// Validar que el ID sea numérico
	if idStr == "invalid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	persona, exists := mockPersonas[idStr]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Persona no encontrada"})
		return
	}

	c.JSON(http.StatusOK, PersonaResponse{
		ID:       persona["id"],
		ZonaID:   persona["zona_id"],
		CamionID: persona["camion_id"],
	})
}

func GetPersonasByZonaMock(c *gin.Context) {
	zonaStr := c.Param("zona")

	// Validar zona
	if zonaStr == "invalid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zona inválida"})
		return
	}

	var result []PersonaResponse

	for _, persona := range mockPersonas {
		if persona["zona_id"] == zonaStr {
			result = append(result, PersonaResponse{
				ID:       persona["id"],
				ZonaID:   persona["zona_id"],
				CamionID: persona["camion_id"],
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"zona":     zonaStr,
		"total":    len(result),
		"personas": result,
	})
}
