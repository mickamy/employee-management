package repository_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/errors/aerrors"
	"github.com/mickamy/employee-management/internal/feature/organization/fixture"
	"github.com/mickamy/employee-management/internal/feature/organization/model"
	"github.com/mickamy/employee-management/internal/feature/organization/repository"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
	"github.com/mickamy/employee-management/internal/test/tdb"
)

func TestDepartment_Create(t *testing.T) {
	t.Parallel()

	// arrange
	db := tdb.New(t)
	sut := repository.NewDepartment(db.Reader)
	department := fixture.Department()

	// act
	err := db.Transactor.WithTx(t.Context(), func(tx tx.Tx) error {
		return sut.Bind(tx).Create(t.Context(), department)
	})

	// assert
	require.NoError(t, err)
	ms, err := sut.List(t.Context())
	require.NoError(t, err)
	require.Len(t, ms, 1)
	assert.Equal(t, department, ms[0])
}

func TestDepartment_Create_DuplicateCode(t *testing.T) {
	t.Parallel()

	// arrange
	db := tdb.New(t)
	sut := repository.NewDepartment(db.Reader)
	department := fixture.Department()
	duplicate := fixture.Department(func(m *model.Department) {
		m.Code = department.Code
	})
	require.NoError(t, db.Transactor.WithTx(t.Context(), func(tx tx.Tx) error {
		return sut.Bind(tx).Create(t.Context(), department)
	}))

	// act
	err := db.Transactor.WithTx(t.Context(), func(tx tx.Tx) error {
		return sut.Bind(tx).Create(t.Context(), duplicate)
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrConflict)
}

func TestDepartment_Rename(t *testing.T) {
	t.Parallel()

	// arrange
	db := tdb.New(t)
	sut := repository.NewDepartment(db.Reader)
	department := fixture.Department()
	require.NoError(t, db.Transactor.WithTx(t.Context(), func(tx tx.Tx) error {
		return sut.Bind(tx).Create(t.Context(), department)
	}))

	// act
	var renamed model.Department
	err := db.Transactor.WithTx(t.Context(), func(tx tx.Tx) error {
		m, err := sut.Bind(tx).Rename(t.Context(), department.ID, "Renamed")
		renamed = m
		return err
	})

	// assert
	require.NoError(t, err)
	assert.Equal(t, department.ID, renamed.ID)
	assert.Equal(t, department.Code, renamed.Code)
	assert.Equal(t, "Renamed", renamed.Name)
}

func TestDepartment_Rename_NotFound(t *testing.T) {
	t.Parallel()

	// arrange
	db := tdb.New(t)
	sut := repository.NewDepartment(db.Reader)

	// act
	err := db.Transactor.WithTx(t.Context(), func(tx tx.Tx) error {
		_, err := sut.Bind(tx).Rename(t.Context(), uuid.New(), "Renamed")
		return err
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrNotFound)
}

func TestDepartment_List(t *testing.T) {
	t.Parallel()

	// arrange
	db := tdb.New(t)
	sut := repository.NewDepartment(db.Reader)
	// created in the opposite order of code, so the returned order can only
	// come from ORDER BY code.
	second := fixture.Department(func(m *model.Department) { m.Code = "b-team" })
	first := fixture.Department(func(m *model.Department) { m.Code = "a-team" })
	require.NoError(t, db.Transactor.WithTx(t.Context(), func(tx tx.Tx) error {
		bound := sut.Bind(tx)
		if err := bound.Create(t.Context(), second); err != nil {
			return err
		}
		return bound.Create(t.Context(), first)
	}))

	// act
	ms, err := sut.List(t.Context())

	// assert
	require.NoError(t, err)
	require.Len(t, ms, 2)
	assert.Equal(t, first, ms[0])
	assert.Equal(t, second, ms[1])
}
