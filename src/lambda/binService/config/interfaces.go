package config

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Result define una interfaz desacoplada del driver de Neo4j.
// Incluye solo los métodos que realmente usa tu aplicación.
type Result interface {
	Next(ctx context.Context) bool
	Record() *neo4j.Record
	Err() error
	Collect(ctx context.Context) ([]*neo4j.Record, error)
	IsOpen() bool
}

// ManagedTransaction representa una transacción genérica (real o mock).
type ManagedTransaction interface {
	Run(ctx context.Context, query string, params map[string]any) (Result, error)
}

// NeoSession representa una sesión de Neo4j, desacoplada del driver.
type NeoSession interface {
	Close(ctx context.Context) error
	ExecuteWrite(ctx context.Context, work func(ManagedTransaction) (any, error)) (any, error)
}

// ---------------- MySQL / GORM ----------------

// DBExecutor es una interfaz desacoplada de GORM, solo con lo que usamos.
type DBExecutor interface {
	WithContext(ctx context.Context) DBExecutor
	Model(value any) DBExecutor
	Where(query string, args ...any) DBExecutor
	Update(column string, value any) DBExecutor
	Error() error
	First(dest any, conds ...any) DBExecutor
	Raw(sql string, values ...any) DBExecutor
	Scan(dest any) DBExecutor
	Exec(sql string, args ...any) DBResult
}
