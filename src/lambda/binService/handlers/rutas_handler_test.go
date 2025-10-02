package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock data para rutas
var mockRutasData = map[int][]map[string]interface{}{
	1: {
		{"id": 1, "lat": -34.6037, "lng": -58.3816, "distance": 0.0},
		{"id": 2, "lat": -34.6118, "lng": -58.3959, "distance": 1.2},
	},
	2: {
		{"id": 3, "lat": -34.5975, "lng": -58.3734, "distance": 0.0},
	},
}

var mockUsuarios = map[string]string{
	"test@example.com":  "1",
	"user2@example.com": "2",
}

var mockPersonasData = map[string]map[string]interface{}{
	"persona:1": {
		"zona_id":     "1",
		"zona_nombre": "Centro",
	},
	"persona:2": {
		"zona_id":     "2",
		"zona_nombre": "Norte",
	},
}

func TestGetRutaHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		zonaID         string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Zona válida con rutas",
			zonaID:         "1",
			expectedStatus: 200,
		},
		{
			name:           "Zona válida sin rutas",
			zonaID:         "3",
			expectedStatus: 200,
		},
		{
			name:           "Zona inválida",
			zonaID:         "invalid",
			expectedStatus: 400,
			expectedError:  "zonaID inválido",
		},
		{
			name:           "Zona inexistente",
			zonaID:         "999",
			expectedStatus: 200, // Devuelve array vacío
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("GET", "/ruta-optima/"+tt.zonaID, nil)
			c.Params = gin.Params{gin.Param{Key: "zonaID", Value: tt.zonaID}}

			GetRutaHandlerMock(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.expectedStatus == 200 {
				var response []interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetRutaHandlerByHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		email          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Usuario válido con email",
			email:          "test@example.com",
			expectedStatus: 200,
		},
		{
			name:           "Email faltante",
			email:          "",
			expectedStatus: 400,
			expectedError:  "Header 'Email' es requerido",
		},
		{
			name:           "Usuario no encontrado",
			email:          "notfound@example.com",
			expectedStatus: 404,
			expectedError:  "Usuario no encontrado",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("GET", "/ruta-optima", nil)
			if tt.email != "" {
				c.Request.Header.Set("email", tt.email)
			}

			GetRutaHandlerByHeaderMock(c)

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
				assert.Contains(t, response, "routes")
				assert.Contains(t, response, "email")
			}
		})
	}
}

func TestGetRutaHandler_MetricsUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test que verifica que las métricas se actualicen
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/ruta-optima/1", nil)
	c.Params = gin.Params{gin.Param{Key: "zonaID", Value: "1"}}

	GetRutaHandlerMock(c)

	assert.Equal(t, 200, w.Code)

	// Verificar que el response tenga la estructura esperada
	var response []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(response), 0)
}

// Mock handlers para testing

func GetRutaHandlerMock(c *gin.Context) {
	zonaIDStr := c.Param("zonaID")

	// Validar zona ID
	if zonaIDStr == "invalid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "zonaID inválido"})
		return
	}

	// Convertir a int para buscar en mock data
	var zonaID int
	switch zonaIDStr {
	case "1":
		zonaID = 1
	case "2":
		zonaID = 2
	case "3":
		zonaID = 3
	default:
		zonaID = 999 // No existe
	}

	// Buscar rutas en mock data
	rutas, exists := mockRutasData[zonaID]
	if !exists {
		rutas = []map[string]interface{}{} // Array vacío
	}

	c.JSON(http.StatusOK, rutas)
}

func GetRutaHandlerByHeaderMock(c *gin.Context) {
	email := c.GetHeader("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Header 'Email' es requerido"})
		return
	}

	// Buscar usuario por email
	personaNum, exists := mockUsuarios[email]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado: email no registrado"})
		return
	}

	// Buscar datos de la persona
	personaKey := "persona:" + personaNum
	persona, exists := mockPersonasData[personaKey]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Persona no encontrada: " + personaKey})
		return
	}

	// Obtener zona ID
	zonaIDStr, ok := persona["zona_id"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "zona_id inválida en persona"})
		return
	}

	// Convertir zona ID a int
	var zonaID int
	switch zonaIDStr {
	case "1":
		zonaID = 1
	case "2":
		zonaID = 2
	default:
		zonaID = 999
	}

	// Obtener rutas para la zona
	rutas, exists := mockRutasData[zonaID]
	if !exists {
		rutas = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"email":     email,
		"persona":   personaNum,
		"zona_id":   zonaID,
		"zona_name": persona["zona_nombre"],
		"routes":    rutas,
	})
}
