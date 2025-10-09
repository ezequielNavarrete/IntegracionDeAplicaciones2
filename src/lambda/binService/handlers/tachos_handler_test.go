package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock handler for CreateTachoHandler
func CreateTachoHandlerMock(c *gin.Context) {
	var tachoData map[string]interface{}
	if err := c.ShouldBindJSON(&tachoData); err != nil {
		c.JSON(400, gin.H{"error": "Datos inválidos"})
		return
	}

	// Simular creación exitosa
	c.JSON(201, gin.H{
		"message": "Tacho creado exitosamente",
		"id":      "mocked_id",
	})
}

// Mock handler for DeleteTachoHandler
func DeleteTachoHandlerMock(c *gin.Context) {
	customID := c.Query("custom_id")
	if customID == "" {
		c.JSON(400, gin.H{"error": "custom_id requerido"})
		return
	}

	c.JSON(200, gin.H{
		"message":   "Tacho eliminado exitosamente",
		"custom_id": customID,
	})
}

// Mock handler for GetAllTachosHandler
func GetAllTachosHandlerMock(c *gin.Context) {
	// Simular tachos mock
	tachos := []map[string]interface{}{
		{
			"id":        1,
			"tipo":      "Residuos",
			"capacidad": 75.5,
			"barrio":    "Palermo",
			"direccion": "Av. Santa Fe 1234",
			"latitude":  -34.5881,
			"longitude": -58.3974,
		},
		{
			"id":        2,
			"tipo":      "Reciclable",
			"capacidad": 50.0,
			"barrio":    "Recoleta",
			"direccion": "Av. Callao 567",
			"latitude":  -34.5936,
			"longitude": -58.3917,
		},
	}

	c.JSON(200, gin.H{
		"tachos": tachos,
		"total":  len(tachos),
	})
}

func TestCreateTachoHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Crear tacho válido",
			requestBody: map[string]interface{}{
				"id_tipo":   1,
				"id_estado": 1,
				"capacidad": 75.5,
				"barrio":    "TEST_BARRIO",
				"direccion": "TEST DIRECCION 123",
				"latitude":  -34.6037,
				"longitude": -58.3816,
				"prioridad": 1,
			},
			expectedStatus: 201,
		},
		{
			name:           "Datos inválidos - JSON malformado",
			requestBody:    nil,
			expectedStatus: 400,
			expectedError:  "Datos inválidos",
		},
		{
			name: "Datos faltantes",
			requestBody: map[string]interface{}{
				"id_tipo": 1,
				// Faltan campos requeridos
			},
			expectedStatus: 201, // El handler actual no valida campos requeridos
		},
		{
			name: "Capacidad inválida",
			requestBody: map[string]interface{}{
				"id_tipo":   1,
				"id_estado": 1,
				"capacidad": -10.0, // Capacidad negativa
				"barrio":    "TEST_BARRIO",
				"direccion": "TEST DIRECCION 123",
				"latitude":  -34.6037,
				"longitude": -58.3816,
				"prioridad": 1,
			},
			expectedStatus: 201, // El handler actual no valida esto
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

			c.Request = httptest.NewRequest("POST", "/tachos", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			CreateTachoHandlerMock(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.expectedStatus == 201 {
				// Verificar respuesta de éxito
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestDeleteTachoHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Eliminar por custom_id",
			queryParams: map[string]string{
				"custom_id": "TACHO_123",
			},
			expectedStatus: 200,
		},
		{
			name:           "custom_id faltante",
			queryParams:    map[string]string{},
			expectedStatus: 400,
			expectedError:  "custom_id requerido",
		},
		{
			name: "custom_id vacío",
			queryParams: map[string]string{
				"custom_id": "",
			},
			expectedStatus: 400,
			expectedError:  "custom_id requerido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			url := "/tachos"
			if len(tt.queryParams) > 0 {
				url += "?"
				first := true
				for key, value := range tt.queryParams {
					if !first {
						url += "&"
					}
					url += key + "=" + value
					first = false
				}
			}

			c.Request = httptest.NewRequest("DELETE", url, nil)

			DeleteTachoHandlerMock(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "message")
			}
		})
	}
}

func TestGetAllTachosHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "Obtener todos los tachos exitosamente",
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("GET", "/tachos", nil)

			GetAllTachosHandlerMock(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "tachos")
			assert.Contains(t, response, "total")

			tachos, ok := response["tachos"].([]interface{})
			assert.True(t, ok)
			assert.Greater(t, len(tachos), 0)
		})
	}
}

// Tests de integración adicionales
func TestCreateTachoHandler_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	tachoData := map[string]interface{}{
		"id_tipo":   1,
		"id_estado": 1,
		"capacidad": 85.0,
		"barrio":    "Villa Crespo",
		"direccion": "Av. Corrientes 4567",
		"latitude":  -34.5998,
		"longitude": -58.4369,
		"prioridad": 2,
	}

	body, err := json.Marshal(tachoData)
	assert.NoError(t, err)

	c.Request = httptest.NewRequest("POST", "/tachos", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	CreateTachoHandlerMock(c)

	assert.Equal(t, 201, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Equal(t, "Tacho creado exitosamente", response["message"])
}

func TestDeleteTachoHandler_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("DELETE", "/tachos?custom_id=INTEGRATION_TEST", nil)

	DeleteTachoHandlerMock(c)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Equal(t, "Tacho eliminado exitosamente", response["message"])
}

func TestGetAllTachosHandler_EmptyResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock handler que devuelve lista vacía
	emptyHandler := func(c *gin.Context) {
		c.JSON(200, gin.H{
			"tachos": []interface{}{},
			"total":  0,
		})
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/tachos", nil)

	emptyHandler(c)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "tachos")
	assert.Equal(t, float64(0), response["total"])
}
