package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/errors/aerrors"
	"github.com/mickamy/employee-management/internal/feature/organization/fixture"
	"github.com/mickamy/employee-management/internal/feature/organization/usecase"
	"github.com/mickamy/employee-management/internal/test/tinfra"
)

func TestCreateDepartment_Do(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	sut := usecase.NewCreateDepartment(infra)
	department := fixture.Department()

	// act
	out, err := sut.Do(t.Context(), usecase.CreateDepartmentInput{
		Code: department.Code,
		Name: department.Name,
	})

	// assert
	require.NoError(t, err)
	assert.NotEmpty(t, out.Department.ID)
	assert.Equal(t, department.Code, out.Department.Code)
	assert.Equal(t, department.Name, out.Department.Name)
}

func TestCreateDepartment_Do_DuplicateCode(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	sut := usecase.NewCreateDepartment(infra)
	department := fixture.Department()
	_, err := sut.Do(t.Context(), usecase.CreateDepartmentInput{
		Code: department.Code,
		Name: department.Name,
	})
	require.NoError(t, err)

	// act
	_, err = sut.Do(t.Context(), usecase.CreateDepartmentInput{
		Code: department.Code,
		Name: department.Name,
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrConflict)
}
