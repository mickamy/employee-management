package usecase_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/errors/aerrors"
	"github.com/mickamy/employee-management/internal/feature/organization/fixture"
	"github.com/mickamy/employee-management/internal/feature/organization/repository"
	"github.com/mickamy/employee-management/internal/feature/organization/usecase"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
	"github.com/mickamy/employee-management/internal/test/tinfra"
)

func TestRenameDepartment_Do(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	department := fixture.Department()
	require.NoError(t, infra.Transactor.WithTx(t.Context(), func(tx tx.Tx) error {
		return repository.NewDepartment(infra.Reader).Bind(tx).Create(t.Context(), department)
	}))
	sut := usecase.NewRenameDepartment(infra)

	// act
	out, err := sut.Do(t.Context(), usecase.RenameDepartmentInput{
		ID:   department.ID,
		Name: "Renamed",
	})

	// assert
	require.NoError(t, err)
	assert.Equal(t, department.ID, out.Department.ID)
	assert.Equal(t, department.Code, out.Department.Code)
	assert.Equal(t, "Renamed", out.Department.Name)
}

func TestRenameDepartment_Do_NotFound(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	sut := usecase.NewRenameDepartment(infra)

	// act
	_, err := sut.Do(t.Context(), usecase.RenameDepartmentInput{
		ID:   uuid.New(),
		Name: "Renamed",
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrNotFound)
}
