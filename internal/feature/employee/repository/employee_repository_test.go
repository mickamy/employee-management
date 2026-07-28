package repository_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/errors/aerrors"
	"github.com/mickamy/employee-management/internal/feature/employee/fixture"
	"github.com/mickamy/employee-management/internal/feature/employee/repository"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
	"github.com/mickamy/employee-management/internal/test/tdb"
)

func TestEmployee_Find(t *testing.T) {
	t.Parallel()

	// arrange
	db := tdb.New(t)
	sut := repository.NewEmployee(db.Reader)
	employee, hire := fixture.EmployeeAndHire()
	require.NoError(t, tx.NewTransactor(db.Writer).WithTx(t.Context(), func(tx tx.Tx) error {
		bound := sut.Bind(tx)
		if err := bound.Create(t.Context(), employee); err != nil {
			return err
		}
		return bound.CreateHire(t.Context(), hire)
	}))

	// act
	m, err := sut.Find(t.Context(), employee.ID)

	// assert
	require.NoError(t, err)
	assert.Equal(t, employee, m)
}

func TestEmployee_Find_NotFound(t *testing.T) {
	t.Parallel()

	// arrange
	db := tdb.New(t)
	sut := repository.NewEmployee(db.Reader)

	// act
	m, err := sut.Find(t.Context(), uuid.New())

	// assert
	require.Empty(t, m)
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrNotFound)
}

func TestEmployee_List(t *testing.T) {
	t.Parallel()

	// arrange
	db := tdb.New(t)
	sut := repository.NewEmployee(db.Reader)
	employee, hire := fixture.EmployeeAndHire()
	require.NoError(t, tx.NewTransactor(db.Writer).WithTx(t.Context(), func(tx tx.Tx) error {
		bound := sut.Bind(tx)
		if err := bound.Create(t.Context(), employee); err != nil {
			return err
		}
		return bound.CreateHire(t.Context(), hire)
	}))

	// act
	ms, err := sut.List(t.Context(), uuid.Nil, 1)

	// assert
	require.NoError(t, err)
	require.Len(t, ms, 1)
	assert.Equal(t, employee, ms[0])
}

func TestEmployee_CreateAndCreateHire(t *testing.T) {
	t.Parallel()

	// arrange
	db := tdb.New(t)
	sut := repository.NewEmployee(db.Reader)
	employee, hire := fixture.EmployeeAndHire()

	// act
	err := tx.NewTransactor(db.Writer).WithTx(t.Context(), func(tx tx.Tx) error {
		bound := sut.Bind(tx)
		if err := bound.Create(t.Context(), employee); err != nil {
			return err
		}
		return bound.CreateHire(t.Context(), hire)
	})

	// assert
	require.NoError(t, err)
	found, err := sut.Find(t.Context(), employee.ID)
	require.NoError(t, err)
	assert.Equal(t, employee, found)
}
