// Package migrate applies the embedded SQL migrations programmatically. It
// exists for tests that need to bring throwaway databases to the current
// schema; deployments run migrations through the goose CLI (make db-migrate),
// and the server binary never links this package.
package migrate

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
)

//go:embed sql/*.sql
var migrations embed.FS

// Up applies all pending migrations through the given admin pool.
func Up(ctx context.Context, pool *pgxpool.Pool) error {
	fsys, err := fs.Sub(migrations, "sql")
	if err != nil {
		return fmt.Errorf("sub filesystem: %w", err)
	}
	provider, err := goose.NewProvider(database.DialectPostgres, stdlib.OpenDBFromPool(pool), fsys)
	if err != nil {
		return fmt.Errorf("create goose provider: %w", err)
	}
	defer func(provider *goose.Provider) {
		_ = provider.Close()
	}(provider)

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
