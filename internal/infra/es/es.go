// Package es is the event store: it appends and reads INSERT-only event tables. Each context
// owns one table; the store is bound to it by name. Optimistic concurrency
// comes from the UNIQUE (stream_id, stream_revision) constraint.
package es

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/mickamy/employee-management/internal/infra/storage/db"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
)

// ErrRevisionConflict reports that another writer appended to the stream
// after it was loaded.
var ErrRevisionConflict = errors.New("stream revision conflict")

// Record is a stored event.
type Record struct {
	GlobalPosition int64
	StreamID       uuid.UUID
	StreamRevision int64
	Type           string
	Payload        []byte
	PersistedAt    time.Time
}

// Unsaved is an event to append.
type Unsaved struct {
	Type    string
	Payload []byte
}

// Store addresses one context's event table. The table name is a
// compile-time constant owned by the feature, never user input.
type Store struct {
	table string
}

func NewStore(table string) Store {
	return Store{table: table}
}

func (s Store) Bind(tx tx.Tx) Bound {
	return s.With(tx.DBTX())
}

// With binds the store to any DBTX, e.g., a read pool for log reads.
func (s Store) With(db tx.DBTX) Bound {
	return Bound{table: s.table, db: db}
}

// Bound is a Store bound to a transaction.
type Bound struct {
	table string
	db    tx.DBTX
}

// Load returns a stream's events in revision order.
func (b Bound) Load(ctx context.Context, streamID uuid.UUID) ([]Record, error) {
	query := fmt.Sprintf(
		`SELECT global_position, stream_id, stream_revision, event_type, payload, persisted_at
		 FROM %s WHERE stream_id = $1 ORDER BY stream_revision`,
		pgx.Identifier{b.table}.Sanitize(),
	)
	rows, err := b.db.Query(ctx, query, streamID)
	if err != nil {
		return nil, fmt.Errorf("load stream: %w", err)
	}
	records, err := pgx.CollectRows(rows, pgx.RowToStructByPos[Record])
	if err != nil {
		return nil, fmt.Errorf("collect stream: %w", err)
	}
	return records, nil
}

// Append writes events after lastRevision and returns the global position of
// the last one. It fails with ErrRevisionConflict when the stream has moved.
func (b Bound) Append(ctx context.Context, streamID uuid.UUID, lastRevision int64, events ...Unsaved) (int64, error) {
	query := fmt.Sprintf(
		`INSERT INTO %s (stream_id, stream_revision, event_type, payload)
		 VALUES ($1, $2, $3, $4) RETURNING global_position`,
		pgx.Identifier{b.table}.Sanitize(),
	)

	var position int64
	for i, e := range events {
		err := b.db.QueryRow(ctx, query, streamID, lastRevision+int64(i)+1, e.Type, e.Payload).Scan(&position)
		if db.IsUniqueViolation(err) {
			return 0, ErrRevisionConflict
		}
		if err != nil {
			return 0, fmt.Errorf("append event: %w", err)
		}
	}
	return position, nil
}

// Read returns up to limit events with a global position greater than
// after, in position order. Rebuild and outbox tooling page through the log with it.
func (b Bound) Read(ctx context.Context, after int64, limit int32) ([]Record, error) {
	query := fmt.Sprintf(
		`SELECT global_position, stream_id, stream_revision, event_type, payload, persisted_at
		 FROM %s WHERE global_position > $1 ORDER BY global_position LIMIT $2`,
		pgx.Identifier{b.table}.Sanitize(),
	)
	rows, err := b.db.Query(ctx, query, after, limit)
	if err != nil {
		return nil, fmt.Errorf("read events: %w", err)
	}
	records, err := pgx.CollectRows(rows, pgx.RowToStructByPos[Record])
	if err != nil {
		return nil, fmt.Errorf("collect events: %w", err)
	}
	return records, nil
}
