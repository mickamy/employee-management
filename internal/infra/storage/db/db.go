// Package db provides the PostgreSQL connection pool shared by features.
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// New opens a connection pool for url and verifies connectivity.
func New(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// Writer is the pool for the command side and projectors, connected as a role
// that can write. The type keeps write and read pools apart at compile time;
// what each role may do is enforced by database grants.
type Writer struct {
	*pgxpool.Pool
}

// Reader is the pool for the query side, connected as a role limited to SELECT.
type Reader struct {
	*pgxpool.Pool
}

// NewWriter opens the write-side pool.
func NewWriter(ctx context.Context, url string) (Writer, error) {
	pool, err := New(ctx, url)
	if err != nil {
		return Writer{}, err
	}
	return Writer{Pool: pool}, nil
}

// NewReader opens the read-side pool.
func NewReader(ctx context.Context, url string) (Reader, error) {
	pool, err := New(ctx, url)
	if err != nil {
		return Reader{}, err
	}
	return Reader{Pool: pool}, nil
}
