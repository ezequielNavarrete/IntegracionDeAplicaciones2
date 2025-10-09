package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mock Tachos (reemplaza MySQL)
var mockTachos = map[int64]string{
	1: "neo1", // IDTacho: IDNeo
	2: "neo2",
	3: "neo3",
}

// Handler minimalista que no toca Neo4j ni MySQL
func UpdatePrioridadTachoHandlerMinimal(c *gin.Context) {
	// ID del tacho
	idStr := c.Param("id_tacho")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Body
	var body UpdatePrioridadRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	// "Buscar" tacho en mock DB
	neoID, ok := mockTachos[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tacho no encontrado"})
		return
	}

	// Actualizamos "prioridad" (simulado)
	c.JSON(http.StatusOK, UpdatePrioridadResponse{
		Message:   "Prioridad actualizada correctamente",
		IDTacho:   id,
		IDNeo:     neoID,
		Prioridad: body.Prioridad,
	})
}

func TestUpdatePrioridadTachoHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		tachoID        string
		requestBody    UpdatePrioridadRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "Actualización exitosa",
			tachoID: "1",
			requestBody: UpdatePrioridadRequest{
				Prioridad: 3,
			},
			expectedStatus: 200,
		},
		{
			name:    "Prioridad mínima",
			tachoID: "1",
			requestBody: UpdatePrioridadRequest{
				Prioridad: 1,
			},
			expectedStatus: 200,
		},
		{
			name:    "Prioridad máxima",
			tachoID: "1",
			requestBody: UpdatePrioridadRequest{
				Prioridad: 5,
			},
			expectedStatus: 200,
		},
		{
			name:    "Tacho existente diferente",
			tachoID: "2",
			requestBody: UpdatePrioridadRequest{
				Prioridad: 2,
			},
			expectedStatus: 200,
		},
		{
			name:           "ID inválido",
			tachoID:        "invalid",
			requestBody:    UpdatePrioridadRequest{Prioridad: 3},
			expectedStatus: 400,
			expectedError:  "ID inválido",
		},
		{
			name:           "Tacho no encontrado",
			tachoID:        "999",
			requestBody:    UpdatePrioridadRequest{Prioridad: 3},
			expectedStatus: 404,
			expectedError:  "Tacho no encontrado",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			c.Request = httptest.NewRequest("PUT", "/tachos/"+tt.tachoID+"/prioridad", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{gin.Param{Key: "id_tacho", Value: tt.tachoID}}

			UpdatePrioridadTachoHandlerMinimal(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response UpdatePrioridadResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Prioridad actualizada correctamente", response.Message)
				assert.Equal(t, tt.requestBody.Prioridad, response.Prioridad)
				assert.NotEmpty(t, response.IDNeo)
			}
		})
	}
}

func TestUpdatePrioridadTachoHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// JSON inválido
	invalidJSON := []byte(`{"prioridad": "invalid"}`)
	
	c.Request = httptest.NewRequest("PUT", "/tachos/1/prioridad", bytes.NewBuffer(invalidJSON))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id_tacho", Value: "1"}}

	UpdatePrioridadTachoHandlerMinimal(c)

	assert.Equal(t, 400, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Datos inválidos")
}

func TestUpdatePrioridadTachoHandler_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Body vacío
	c.Request = httptest.NewRequest("PUT", "/tachos/1/prioridad", bytes.NewBuffer([]byte("{}")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id_tacho", Value: "1"}}

	UpdatePrioridadTachoHandlerMinimal(c)

	assert.Equal(t, 200, w.Code)
	
	var response UpdatePrioridadResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Prioridad) // Valor por defecto
}

func TestUpdatePrioridadTachoHandler_DifferentPriorities(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	priorities := []int{1, 2, 3, 4, 5, 0, -1, 10}
	
	for _, priority := range priorities {
		t.Run(fmt.Sprintf("Prioridad_%d", priority), func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			requestBody := UpdatePrioridadRequest{Prioridad: priority}
			body, _ := json.Marshal(requestBody)

			c.Request = httptest.NewRequest("PUT", "/tachos/1/prioridad", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{gin.Param{Key: "id_tacho", Value: "1"}}

			UpdatePrioridadTachoHandlerMinimal(c)

			assert.Equal(t, 200, w.Code)
			
			var response UpdatePrioridadResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, priority, response.Prioridad)
		})
	}
}