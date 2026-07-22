// Package tx provides use-case-owned transaction boundaries. It exposes only
// the transaction vocabulary; connection pools stay in the db package.
package tx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/mickamy/employee-management/internal/storage/db"
)

// DBTX is the narrow query interface repositories hand to sqlc. It matches
// the interface sqlc generates with sql_package pgx/v5 and is satisfied by
// both pools and transactions.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Tx exposes a transaction as DBTX. Commit and rollback stay with the
// Transactor, so the transaction boundary remains the use case's.
type Tx struct {
	tx pgx.Tx
}

// DBTX exposes the transaction as the narrow query interface.
func (t Tx) DBTX() DBTX { return t.tx }

// Transactor runs a function within a single write transaction.
type Transactor interface {
	WithTx(ctx context.Context, fn func(tx Tx) error) error
}

type transactor struct {
	writer db.Writer
}

var _ Transactor = transactor{}

func NewTransactor(writer db.Writer) Transactor {
	return transactor{writer: writer}
}

func (t transactor) WithTx(ctx context.Context, fn func(tx Tx) error) error {
	pgxTx, err := t.writer.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	return run(ctx, pgxTx, fn)
}

// ReadTransactor runs a function within a read-only transaction on the reader
// pool, giving multiple reads one consistent snapshot.
type ReadTransactor interface {
	WithReadTx(ctx context.Context, fn func(tx Tx) error) error
}

type readTransactor struct {
	reader db.Reader
}

var _ ReadTransactor = readTransactor{}

func NewReadTransactor(reader db.Reader) ReadTransactor {
	return readTransactor{reader: reader}
}

func (t readTransactor) WithReadTx(ctx context.Context, fn func(tx Tx) error) error {
	pgxTx, err := t.reader.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return fmt.Errorf("begin read-only: %w", err)
	}
	return run(ctx, pgxTx, fn)
}

func run(ctx context.Context, pgxTx pgx.Tx, fn func(tx Tx) error) error {
	defer func() {
		_ = pgxTx.Rollback(ctx) // no-op if already committed
	}()

	if err := fn(Tx{tx: pgxTx}); err != nil {
		return err
	}
	if err := pgxTx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
