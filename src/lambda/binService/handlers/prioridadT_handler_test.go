package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Test simplificado para la nueva funcionalidad de características
func TestUpdatePrioridadTachoHandler_NewStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PUT("/tachos/:id_tacho/prioridad", UpdatePrioridadTachoHandler)

	tests := []struct {
		name     string
		id       string
		request  services.UpdateTachoCaracteristicasRequest
		wantCode int
	}{
		{
			"Happy path - actualizar una característica",
			"1",
			services.UpdateTachoCaracteristicasRequest{
				Caracteristicas: []services.UpdateCaracteristicaRequest{
					{Nombre: "Humedad", Prioridad: 3},
				},
			},
			200,
		},
		{
			"Happy path - actualizar múltiples características",
			"1",
			services.UpdateTachoCaracteristicasRequest{
				Caracteristicas: []services.UpdateCaracteristicaRequest{
					{Nombre: "Humedad", Prioridad: 4},
					{Nombre: "Olor", Prioridad: 2},
					{Nombre: "Llenado", Prioridad: 3},
				},
			},
			200,
		},
		{
			"ID inválido",
			"abc",
			services.UpdateTachoCaracteristicasRequest{
				Caracteristicas: []services.UpdateCaracteristicaRequest{
					{Nombre: "Humedad", Prioridad: 2},
				},
			},
			400,
		},
		{
			"Sin características",
			"1",
			services.UpdateTachoCaracteristicasRequest{
				Caracteristicas: []services.UpdateCaracteristicaRequest{},
			},
			400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest(http.MethodPut, "/tachos/"+tt.id+"/prioridad", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}
