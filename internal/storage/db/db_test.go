package db_test

import (
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mickamy/employee-management/internal/storage/db"
)

func TestNew(t *testing.T) {
	t.Parallel()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL is not set")
	}

	pool, err := db.New(t.Context(), url)
	if err != nil {
		t.Fatalf("db.New() error = %v", err)
	}
	pool.Close()
}

func TestNewWriter(t *testing.T) {
	t.Parallel()

	url := os.Getenv("DATABASE_WRITER_URL")
	if url == "" {
		t.Skip("DATABASE_WRITER_URL is not set")
	}

	w, err := db.NewWriter(t.Context(), url)
	if err != nil {
		t.Fatalf("db.NewWriter() error = %v", err)
	}
	defer w.Close()

	assertCurrentUser(t, w.Pool, "app_writer")
}

func TestNewReader(t *testing.T) {
	t.Parallel()

	url := os.Getenv("DATABASE_READER_URL")
	if url == "" {
		t.Skip("DATABASE_READER_URL is not set")
	}

	r, err := db.NewReader(t.Context(), url)
	if err != nil {
		t.Fatalf("db.NewReader() error = %v", err)
	}
	defer r.Close()

	assertCurrentUser(t, r.Pool, "app_reader")
}

func assertCurrentUser(t *testing.T, pool *pgxpool.Pool, want string) {
	t.Helper()

	var got string
	if err := pool.QueryRow(t.Context(), "SELECT current_user").Scan(&got); err != nil {
		t.Fatalf("query current_user: %v", err)
	}
	if got != want {
		t.Fatalf("current_user = %q, want %q", got, want)
	}
}

func TestNewUnreachable(t *testing.T) {
	t.Parallel()

	url := "postgres://127.0.0.1:1/nowhere?connect_timeout=1"
	if _, err := db.New(t.Context(), url); err == nil {
		t.Fatal("want error, got nil")
	}
}
