// Package migrate applies the embedded SQL migrations programmatically. It
// exists for tests that need to bring throwaway databases to the current
// schema; deployments run migrations through the goose CLI (make tdb-migrate),
// and the server binary never links this package.
package migrate

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
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
func Up(ctx context.Context, pool *pgxpool.Pool) (err error) {
	sqldb := stdlib.OpenDBFromPool(pool)
	defer func() {
		err = errors.Join(err, sqldb.Close())
	}()

	err = UpDB(ctx, sqldb)
	return err
}

// UpDB applies all pending migrations through the given database handle,
// which stays open and owned by the caller.
func UpDB(ctx context.Context, sqldb *sql.DB) (err error) {
	fsys, err := fs.Sub(migrations, "sql")
	if err != nil {
		return fmt.Errorf("sub filesystem: %w", err)
	}
	provider, err := goose.NewProvider(database.DialectPostgres, sqldb, fsys)
	if err != nil {
		return fmt.Errorf("create goose provider: %w", err)
	}
	defer func(provider *goose.Provider) {
		err = errors.Join(err, provider.Close())
	}(provider)
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// Hash returns a digest of the embedded migrations for tools that cache
// migrated database templates.
func Hash() (string, error) {
	entries, err := fs.ReadDir(migrations, "sql")
	if err != nil {
		return "", fmt.Errorf("read migrations: %w", err)
	}
	h := sha256.New()
	for _, entry := range entries {
		content, err := migrations.ReadFile("sql/" + entry.Name())
		if err != nil {
			return "", fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		_, _ = h.Write([]byte(entry.Name()))
		_, _ = h.Write(content)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
