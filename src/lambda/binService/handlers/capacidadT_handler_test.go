package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- MOCK DB IMPLEMENTATION COMPATIBLE ---

type mockExecutor struct {
	lastCapacidad float64
}

func (m *mockExecutor) WithContext(ctx context.Context) config.DBExecutor {
	return m
}

func (m *mockExecutor) Model(value any) config.DBExecutor {
	return m
}

func (m *mockExecutor) Where(query string, args ...any) config.DBExecutor {
	return m
}

func (m *mockExecutor) Update(column string, value any) config.DBExecutor {
	if column == "capacidad" {
		if v, ok := value.(float64); ok {
			m.lastCapacidad = v
		}
	}
	return m
}

func (m *mockExecutor) First(dest any, conds ...any) config.DBExecutor {
	return m
}

func (m *mockExecutor) Raw(sql string, values ...any) config.DBExecutor {
	return m
}

func (m *mockExecutor) Scan(dest any) config.DBExecutor {
	return m
}

func (m *mockExecutor) Exec(sql string, args ...any) config.DBResult {
	return &mockResult{}
}

func (m *mockExecutor) Error() error {
	return nil
}

// Mock de DBResult (para Exec)
type mockResult struct{}

func (r *mockResult) Error() error {
	return nil
}

func (r *mockResult) RowsAffected() int64 {
	return 1
}

// --- TEST DEL HANDLER REAL ---

func TestUpdateCapacidadTachoHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock := &mockExecutor{}
	config.DB = mock // ✅ inyectamos mock en el paquete global config

	router := gin.Default()
	router.PUT("/tachos/:id_tacho/capacidad", UpdateCapacidadTachoHandler)

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
			body := UpdateCapacidadRequest{Capacidad: tt.capacidad}
			jsonBody, _ := json.Marshal(body)

			req, _ := http.NewRequest(http.MethodPut, "/tachos/"+tt.id+"/capacidad", bytes.NewBuffer(jsonBody))
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
				assert.Equal(t, tt.capacidad, mock.lastCapacidad)
			}
		})
	}
}
