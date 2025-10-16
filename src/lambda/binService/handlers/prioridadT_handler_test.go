package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mock Tachos (reemplaza MySQL/GORM)
var mockTachos = map[int64]models.Tacho{
	1: {IDTacho: 1, IDNeo: "neo1"},
}

// Handler de test que inyecta mocks
func UpdatePrioridadTachoHandlerTest(c *gin.Context) {
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

	// Mock GORM
	tacho, ok := mockTachos[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tacho no encontrado"})
		return
	}

	// Mock Neo4j usando MockNeoSession
	mockSession := MockNeoSession{}
	_, err = mockSession.ExecuteWrite(context.Background(), func(tx config.ManagedTransaction) (any, error) {
		// simulamos el update, no hacemos nada
		return nil, nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar prioridad en Neo4j"})
		return
	}

	c.JSON(http.StatusOK, UpdatePrioridadResponse{
		Message:   "Prioridad actualizada correctamente",
		IDTacho:   tacho.IDTacho,
		IDNeo:     tacho.IDNeo,
		Prioridad: body.Prioridad,
	})
}

func TestUpdatePrioridadTachoHandler_WithMocks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PUT("/tachos/:id_tacho/prioridad", UpdatePrioridadTachoHandlerTest)

	tests := []struct {
		name       string
		id         string
		prioridad  int
		wantCode   int
		wantErrMsg string
	}{
		{"Happy path", "1", 2, 200, ""},
		{"ID inválido", "abc", 2, 400, "ID inválido"},
		{"Tacho no encontrado", "999", 2, 404, "Tacho no encontrado"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(UpdatePrioridadRequest{Prioridad: tt.prioridad})
			req, _ := http.NewRequest(http.MethodPut, "/tachos/"+tt.id+"/prioridad", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantErrMsg != "" {
				var resp map[string]string
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, tt.wantErrMsg, resp["error"])
			} else {
				var resp UpdatePrioridadResponse
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, tt.id, strconv.FormatInt(resp.IDTacho, 10))
				assert.Equal(t, mockTachos[resp.IDTacho].IDNeo, resp.IDNeo)
				assert.Equal(t, tt.prioridad, resp.Prioridad)
				assert.Equal(t, "Prioridad actualizada correctamente", resp.Message)
			}
		})
	}
}
