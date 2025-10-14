// En config/interfaces.go (por ejemplo)
package config

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type ManagedTransaction interface {
	Run(ctx context.Context, query string, params map[string]any) (neo4j.ResultWithContext, error)
}

type NeoSession interface {
	Close(ctx context.Context) error
	ExecuteWrite(ctx context.Context, work func(ManagedTransaction) (any, error)) (any, error)
}
