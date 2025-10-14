// mocks.go
package handlers

import (
	"context"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Mocks para tests
type MockManagedTx struct{}

func (m MockManagedTx) Run(ctx context.Context, query string, params map[string]any) (neo4j.ResultWithContext, error) {
	// Retornamos un ResultWithContext vacío mockeado
	return nil, nil
}

type MockNeoSession struct{}

func (m MockNeoSession) Close(ctx context.Context) error {
	return nil
}

func (m MockNeoSession) ExecuteWrite(ctx context.Context, work func(tx config.ManagedTransaction) (any, error)) (any, error) {
	return work(MockManagedTx{})
}
