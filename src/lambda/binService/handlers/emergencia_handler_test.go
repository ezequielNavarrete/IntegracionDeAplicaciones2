package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSendEmergencyHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Emergencia válida",
			requestBody: map[string]interface{}{
				"tipo":        "incendio",
				"zona":        "zona_centro",
				"descripcion": "Incendio en edificio",
			},
			expectedStatus: 200,
		},
		{
			name:           "Cuerpo inválido - JSON malformado",
			requestBody:    nil, // Esto causará un error de bind
			expectedStatus: 400,
			expectedError:  "Datos inválidos",
		},
		{
			name: "Emergencia con campos vacíos",
			requestBody: map[string]interface{}{
				"tipo":        "",
				"zona":        "zona_centro",
				"descripcion": "Descripción",
			},
			expectedStatus: 200, // Asumiendo que acepta campos vacíos
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var body []byte
			var err error

			if tt.requestBody != nil {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			} else {
				body = []byte("invalid json")
			}

			c.Request = httptest.NewRequest("POST", "/enviar-emergencia", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			SendEmergencyHandler(c)

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

func TestSendEmergencyHandler_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test de integración completo
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	emergencyData := map[string]interface{}{
		"tipo":        "robo",
		"zona":        "zona_norte",
		"descripcion": "Robo en comercio",
	}

	body, _ := json.Marshal(emergencyData)
	c.Request = httptest.NewRequest("POST", "/enviar-emergencia", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	SendEmergencyHandler(c)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
}
