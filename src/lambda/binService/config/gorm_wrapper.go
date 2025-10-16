package config

import (
	"context"

	"gorm.io/gorm"
)

// wrapper para *gorm.DB
type gormWrapper struct {
	db *gorm.DB
}

// Con puntero receptor
func (g *gormWrapper) WithContext(ctx context.Context) DBExecutor {
	return &gormWrapper{db: g.db.WithContext(ctx)}
}

func (g *gormWrapper) Model(value any) DBExecutor {
	return &gormWrapper{db: g.db.Model(value)}
}

func (g *gormWrapper) Where(query string, args ...any) DBExecutor {
	return &gormWrapper{db: g.db.Where(query, args...)}
}

func (g *gormWrapper) Update(column string, value any) DBExecutor {
	return &gormWrapper{db: g.db.Update(column, value)}
}

func (g *gormWrapper) First(dest any, conds ...any) DBExecutor {
	return &gormWrapper{db: g.db.First(dest, conds...)}
}

// Devuelve el error final de la chain
func (g *gormWrapper) Error() error {
	return g.db.Error
}

func (g *gormWrapper) Raw(sql string, values ...any) DBExecutor {
	return &gormWrapper{db: g.db.Raw(sql, values...)}
}

func (g *gormWrapper) Scan(dest any) DBExecutor {
	return &gormWrapper{db: g.db.Scan(dest)}
}

type DBResult interface {
	Error() error
	RowsAffected() int64
}

func (g *gormWrapper) Exec(sql string, args ...any) DBResult {
	res := g.db.Exec(sql, args...)
	return &gormResultWrapper{res}
}

type gormResultWrapper struct {
	res *gorm.DB
}

func (r *gormResultWrapper) Error() error {
	return r.res.Error
}

func (r *gormResultWrapper) RowsAffected() int64 {
	return r.res.RowsAffected
}
