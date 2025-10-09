package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mock DB para tests
var mockDB = map[int64]float64{
	1: 50,
	2: 75,
}

// Handler minimalista con validación de 0-100
func UpdateCapacidadTachoHandlerMinimal(c *gin.Context) {
	idStr := c.Param("id_tacho")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var body UpdateCapacidadRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	// Validación de capacidad
	if body.Capacidad < 0 || body.Capacidad > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Capacidad fuera de rango"})
		return
	}

	// Actualizar "DB"
	mockDB[id] = body.Capacidad

	c.JSON(http.StatusOK, UpdateCapacidadResponse{
		Message:   "Capacidad actualizada correctamente",
		IDTacho:   id,
		Capacidad: body.Capacidad,
	})
}

func TestUpdateCapacidadTachoHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		tachoID        string
		requestBody    UpdateCapacidadRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "Actualización exitosa",
			tachoID: "1",
			requestBody: UpdateCapacidadRequest{
				Capacidad: 75.0,
			},
			expectedStatus: 200,
		},
		{
			name:    "Capacidad en el límite inferior",
			tachoID: "1",
			requestBody: UpdateCapacidadRequest{
				Capacidad: 0.0,
			},
			expectedStatus: 200,
		},
		{
			name:    "Capacidad en el límite superior",
			tachoID: "1",
			requestBody: UpdateCapacidadRequest{
				Capacidad: 100.0,
			},
			expectedStatus: 200,
		},
		{
			name:    "Capacidad fuera de rango - negativa",
			tachoID: "1",
			requestBody: UpdateCapacidadRequest{
				Capacidad: -10.0,
			},
			expectedStatus: 400,
			expectedError:  "Capacidad fuera de rango",
		},
		{
			name:    "Capacidad fuera de rango - mayor a 100",
			tachoID: "1",
			requestBody: UpdateCapacidadRequest{
				Capacidad: 150.0,
			},
			expectedStatus: 400,
			expectedError:  "Capacidad fuera de rango",
		},
		{
			name:           "ID inválido",
			tachoID:        "invalid",
			requestBody:    UpdateCapacidadRequest{Capacidad: 50.0},
			expectedStatus: 400,
			expectedError:  "ID inválido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			c.Request = httptest.NewRequest("PUT", "/tachos/"+tt.tachoID+"/capacidad", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{gin.Param{Key: "id_tacho", Value: tt.tachoID}}

			UpdateCapacidadTachoHandlerMinimal(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response UpdateCapacidadResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Capacidad actualizada correctamente", response.Message)
				assert.Equal(t, tt.requestBody.Capacidad, response.Capacidad)
			}
		})
	}
}

func TestUpdateCapacidadTachoHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// JSON inválido
	invalidJSON := []byte(`{"capacidad": "invalid"}`)
	
	c.Request = httptest.NewRequest("PUT", "/tachos/1/capacidad", bytes.NewBuffer(invalidJSON))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id_tacho", Value: "1"}}

	UpdateCapacidadTachoHandlerMinimal(c)

	assert.Equal(t, 400, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Datos inválidos")
}

func TestUpdateCapacidadTachoHandler_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name        string
		capacidad   float64
		shouldPass  bool
	}{
		{"Capacidad 0.1", 0.1, true},
		{"Capacidad 99.9", 99.9, true},
		{"Capacidad exactamente 0", 0.0, true},
		{"Capacidad exactamente 100", 100.0, true},
		{"Capacidad -0.1", -0.1, false},
		{"Capacidad 100.1", 100.1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			requestBody := UpdateCapacidadRequest{Capacidad: tt.capacidad}
			body, _ := json.Marshal(requestBody)

			c.Request = httptest.NewRequest("PUT", "/tachos/1/capacidad", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{gin.Param{Key: "id_tacho", Value: "1"}}

			UpdateCapacidadTachoHandlerMinimal(c)

			if tt.shouldPass {
				assert.Equal(t, 200, w.Code)
			} else {
				assert.Equal(t, 400, w.Code)
			}
		})
	}
}

func TestUpdateCapacidadTachoHandlerMinimal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PUT("/tachos/:id_tacho/capacidad", UpdateCapacidadTachoHandlerMinimal)

	tests := []struct {
		name       string
		id         string
		capacidad  float64
		wantCode   int
		wantErrMsg string
	}{
		{"Happy path 50", "1", 50, 200, ""},
		{"Happy path 0", "1", 0, 200, ""},
		{"Happy path 100", "1", 100, 200, ""},
		{"Capacidad negativa", "1", -10, 400, "Capacidad fuera de rango"},
		{"Capacidad demasiado alta", "1", 200, 400, "Capacidad fuera de rango"},
		{"ID inválido", "abc", 50, 400, "ID inválido"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(UpdateCapacidadRequest{Capacidad: tt.capacidad})
			req, _ := http.NewRequest(http.MethodPut, "/tachos/"+tt.id+"/capacidad", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantErrMsg != "" {
				var resp map[string]string
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, tt.wantErrMsg, resp["error"])
			} else {
				var resp UpdateCapacidadResponse
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, tt.capacidad, resp.Capacidad)
				assert.Equal(t, "Capacidad actualizada correctamente", resp.Message)
			}
		})
	}
}
