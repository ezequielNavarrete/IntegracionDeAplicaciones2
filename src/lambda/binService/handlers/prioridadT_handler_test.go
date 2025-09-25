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

// mock Tachos (reemplaza MySQL)
var mockTachos = map[int64]string{
	1: "neo1", // IDTacho: IDNeo
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
	// En este caso solo devolvemos el valor del body

	c.JSON(http.StatusOK, UpdatePrioridadResponse{
		Message:   "Prioridad actualizada correctamente",
		IDTacho:   id,
		IDNeo:     neoID,
		Prioridad: body.Prioridad,
	})
}

func TestUpdatePrioridadTachoHandlerMinimal_Subtests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PUT("/tachos/:id_tacho/prioridad", UpdatePrioridadTachoHandlerMinimal)

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
		// Opcional: límites de prioridad si quisieras validar rangos
		{"Prioridad mínima", "1", 0, 200, ""},
		{"Prioridad máxima", "1", 5, 200, ""},
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
				assert.Equal(t, mockTachos[resp.IDTacho], resp.IDNeo)
				assert.Equal(t, tt.prioridad, resp.Prioridad)
				assert.Equal(t, "Prioridad actualizada correctamente", resp.Message)
			}
		})
	}
}
