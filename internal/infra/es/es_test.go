package es_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/infra/es"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
	"github.com/mickamy/employee-management/internal/test/tdb"
)

// Tests run against the assignment context's event table; the store itself
// is table-agnostic.
const table = "assignment_events"

func TestBound_AppendAndLoad(t *testing.T) {
	t.Parallel()

	// arrange
	db := tdb.New(t)
	store := es.NewStore(table)
	streamID := uuid.New()

	// act
	var position int64
	err := db.Transactor.WithTx(t.Context(), func(tx tx.Tx) error {
		last, err := store.Bind(tx).Append(t.Context(), streamID, 0,
			es.Unsaved{Type: "SomethingHappened", Payload: []byte(`{"n":1}`)},
			es.Unsaved{Type: "SomethingElseHappened", Payload: []byte(`{"n":2}`)},
		)
		position = last
		return err
	})

	// assert
	require.NoError(t, err)
	assert.Positive(t, position)
	records, err := store.With(db.Reader).Load(t.Context(), streamID)
	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, int64(1), records[0].StreamRevision)
	assert.Equal(t, int64(2), records[1].StreamRevision)
	assert.Equal(t, "SomethingHappened", records[0].Type)
	assert.Equal(t, position, records[1].GlobalPosition)
}

func TestBound_Append_RevisionConflict(t *testing.T) {
	t.Parallel()

	// arrange
	db := tdb.New(t)
	store := es.NewStore(table)
	streamID := uuid.New()
	require.NoError(t, db.Transactor.WithTx(t.Context(), func(tx tx.Tx) error {
		_, err := store.Bind(tx).Append(t.Context(), streamID, 0,
			es.Unsaved{Type: "SomethingHappened", Payload: []byte(`{}`)},
		)
		return err
	}))

	// act: append with the same expected revision, as a lost race would
	err := db.Transactor.WithTx(t.Context(), func(tx tx.Tx) error {
		_, err := store.Bind(tx).Append(t.Context(), streamID, 0,
			es.Unsaved{Type: "SomethingHappened", Payload: []byte(`{}`)},
		)
		return err
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, es.ErrRevisionConflict)
}

func TestBound_Read(t *testing.T) {
	t.Parallel()

	// arrange
	db := tdb.New(t)
	store := es.NewStore(table)
	require.NoError(t, db.Transactor.WithTx(t.Context(), func(tx tx.Tx) error {
		_, err := store.Bind(tx).Append(t.Context(), uuid.New(), 0,
			es.Unsaved{Type: "First", Payload: []byte(`{}`)},
			es.Unsaved{Type: "Second", Payload: []byte(`{}`)},
			es.Unsaved{Type: "Third", Payload: []byte(`{}`)},
		)
		return err
	}))

	// act
	all, err := store.With(db.Reader).Read(t.Context(), 0, 10)
	require.NoError(t, err)
	require.Len(t, all, 3)
	rest, err := store.With(db.Reader).Read(t.Context(), all[0].GlobalPosition, 10)

	// assert
	require.NoError(t, err)
	require.Len(t, rest, 2)
	assert.Equal(t, "Second", rest[0].Type)
	assert.Equal(t, "Third", rest[1].Type)
}
