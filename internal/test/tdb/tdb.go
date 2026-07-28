// Package tdb provisions isolated, migrated PostgreSQL databases for tests.
// Each test receives its own database cloned from a migrated template via
// pgtestdb, plus typed pools connected as the writer and reader roles.
package tdb

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/mickamy/employee-management/internal/storage/tx"
	"github.com/peterldowns/pgtestdb"
	"github.com/stretchr/testify/require"

	sdb "github.com/mickamy/employee-management/internal/storage/db"
	"github.com/mickamy/employee-management/internal/storage/db/migrate"
)

// DB is a set of typed pools connected to a freshly migrated database.
type DB struct {
	Writer     sdb.Writer
	Reader     sdb.Reader
	Transactor tx.Transactor
}

// New provisions an isolated database and returns typed pools connected to
// it. Everything is cleaned up with the test. The test is skipped when the
// database URLs are not configured.
func New(t *testing.T) DB {
	t.Helper()

	adminURL := os.Getenv("DATABASE_URL")
	writerURL := os.Getenv("DATABASE_WRITER_URL")
	readerURL := os.Getenv("DATABASE_READER_URL")
	if adminURL == "" || writerURL == "" || readerURL == "" {
		t.Skip("DATABASE_URL, DATABASE_WRITER_URL, and DATABASE_READER_URL must be set")
	}

	cfg := pgtestdb.Custom(t, configFromURL(t, adminURL), migrator{})

	writer, err := sdb.NewWriter(t.Context(), replaceDBName(t, writerURL, cfg.Database))
	require.NoError(t, err)
	t.Cleanup(writer.Close)

	reader, err := sdb.NewReader(t.Context(), replaceDBName(t, readerURL, cfg.Database))
	require.NoError(t, err)
	t.Cleanup(reader.Close)

	return DB{Writer: writer, Reader: reader, Transactor: tx.NewTransactor(writer)}
}

type migrator struct{}

var _ pgtestdb.Migrator = migrator{}

func (migrator) Hash() (string, error) {
	hash, err := migrate.Hash()
	if err != nil {
		return "", fmt.Errorf("hash migrations: %w", err)
	}
	return hash, nil
}

func (migrator) Migrate(ctx context.Context, sqldb *sql.DB, _ pgtestdb.Config) error {
	if err := migrate.UpDB(ctx, sqldb); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

func configFromURL(t *testing.T, raw string) pgtestdb.Config {
	t.Helper()

	u, err := url.Parse(raw)
	require.NoError(t, err)
	password, _ := u.User.Password()
	return pgtestdb.Config{
		DriverName: "pgx",
		Host:       u.Hostname(),
		Port:       u.Port(),
		User:       u.User.Username(),
		Password:   password,
		Database:   strings.TrimPrefix(u.Path, "/"),
		Options:    u.RawQuery,
	}
}

func replaceDBName(t *testing.T, raw, name string) string {
	t.Helper()

	u, err := url.Parse(raw)
	require.NoError(t, err)
	u.Path = "/" + name
	return u.String()
}
