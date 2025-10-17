package handlers

import (
	"context"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// MockManagedTx implementa config.ManagedTransaction
type MockManagedTx struct{}

func (m MockManagedTx) Run(ctx context.Context, query string, params map[string]any) (config.Result, error) {
	return &MockResult{}, nil
}

// MockResult implementa config.Result
type MockResult struct{}

func (m *MockResult) Next(ctx context.Context) bool { return false }
func (m *MockResult) Record() *neo4j.Record         { return nil }
func (m *MockResult) Err() error                    { return nil }
func (m *MockResult) Collect(ctx context.Context) ([]*neo4j.Record, error) {
	return []*neo4j.Record{}, nil
}
func (m *MockResult) IsOpen() bool { return false }

// MockNeoSession implementa config.NeoSession
type MockNeoSession struct{}

func (m MockNeoSession) Close(ctx context.Context) error {
	return nil
}

func (m MockNeoSession) ExecuteWrite(ctx context.Context, work func(tx config.ManagedTransaction) (any, error)) (any, error) {
	return work(MockManagedTx{})
}

func (m MockNeoSession) ExecuteRead(ctx context.Context, work func(tx config.ManagedTransaction) (any, error)) (any, error) {
	return work(MockManagedTx{})
}
