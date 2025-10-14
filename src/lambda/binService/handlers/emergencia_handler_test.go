package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSendEmergencyHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.POST("/enviar-emergencia", SendEmergencyHandler)

	// --- Test JSON válido ---
	t.Run("SendEmergencyHandler - Valid JSON", func(t *testing.T) {
		body := RequestBody{
			Tipo:        "incendio",
			Descripcion: "Incendio en edificio de oficinas",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest(http.MethodPost, "/enviar-emergencia", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)

		assert.Equal(t, "Emergencia enviada correctamente", resp["message"])
		assert.Equal(t, body.Tipo, resp["tipo"])
		assert.Equal(t, body.Descripcion, resp["descripcion"])
	})

	// --- Test JSON inválido ---
	t.Run("SendEmergencyHandler - Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/enviar-emergencia", bytes.NewBuffer([]byte(`{invalid-json`)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)

		assert.Equal(t, "Datos inválidos", resp["error"])
	})
}
