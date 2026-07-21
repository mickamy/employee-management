package db_test

import (
	"os"
	"testing"

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

func TestNewUnreachable(t *testing.T) {
	t.Parallel()

	url := "postgres://127.0.0.1:1/nowhere?connect_timeout=1"
	if _, err := db.New(t.Context(), url); err == nil {
		t.Fatal("want error, got nil")
	}
}
