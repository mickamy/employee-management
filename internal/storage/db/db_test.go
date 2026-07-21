package db_test

import (
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/storage/db"
)

func TestNew(t *testing.T) {
	t.Parallel()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL is not set")
	}

	pool, err := db.New(t.Context(), url)
	require.NoError(t, err)
	pool.Close()
}

func TestNewWriter(t *testing.T) {
	t.Parallel()

	url := os.Getenv("DATABASE_WRITER_URL")
	if url == "" {
		t.Skip("DATABASE_WRITER_URL is not set")
	}

	w, err := db.NewWriter(t.Context(), url)
	require.NoError(t, err)
	defer w.Close()

	requireCurrentUser(t, w.Pool, "app_writer")
}

func TestNewReader(t *testing.T) {
	t.Parallel()

	url := os.Getenv("DATABASE_READER_URL")
	if url == "" {
		t.Skip("DATABASE_READER_URL is not set")
	}

	r, err := db.NewReader(t.Context(), url)
	require.NoError(t, err)
	defer r.Close()

	requireCurrentUser(t, r.Pool, "app_reader")
}

func requireCurrentUser(t *testing.T, pool *pgxpool.Pool, want string) {
	t.Helper()

	var got string
	err := pool.QueryRow(t.Context(), "SELECT current_user").Scan(&got)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestNewUnreachable(t *testing.T) {
	t.Parallel()

	url := "postgres://127.0.0.1:1/nowhere?connect_timeout=1"
	_, err := db.New(t.Context(), url)
	require.Error(t, err)
}
