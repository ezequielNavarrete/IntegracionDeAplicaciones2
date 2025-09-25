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
