package usecase_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/lib/times"

	"github.com/mickamy/employee-management/internal/errors/aerrors"
	"github.com/mickamy/employee-management/internal/feature/assignment/model"
	"github.com/mickamy/employee-management/internal/feature/assignment/usecase"
	"github.com/mickamy/employee-management/internal/test/tinfra"
)

func TestReleaseAssignment_Do(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	employeeID := hireEmployee(t, infra)
	departmentID := createDepartment(t, infra)
	assigned, err := usecase.NewAssignEmployee(infra).Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2026, 4, 1),
	})
	require.NoError(t, err)
	sut := usecase.NewReleaseAssignment(infra)

	// act
	out, err := sut.Do(t.Context(), usecase.ReleaseAssignmentInput{
		AssignmentID: assigned.Assignment.ID,
		ReleasedOn:   times.Date(2026, 9, 30),
	})

	// assert
	require.NoError(t, err)
	assert.Equal(t, assigned.Assignment.ID, out.Assignment.ID)
	require.NotNil(t, out.Assignment.ReleasedOn)
	assert.Equal(t, times.Date(2026, 9, 30), *out.Assignment.ReleasedOn)
}

func TestReleaseAssignment_Do_NotFound(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	sut := usecase.NewReleaseAssignment(infra)

	// act
	_, err := sut.Do(t.Context(), usecase.ReleaseAssignmentInput{
		AssignmentID: uuid.New(),
		ReleasedOn:   times.Date(2026, 9, 30),
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrNotFound)
}

func TestReleaseAssignment_Do_AlreadyReleased(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	employeeID := hireEmployee(t, infra)
	departmentID := createDepartment(t, infra)
	assigned, err := usecase.NewAssignEmployee(infra).Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2026, 4, 1),
	})
	require.NoError(t, err)
	sut := usecase.NewReleaseAssignment(infra)
	_, err = sut.Do(t.Context(), usecase.ReleaseAssignmentInput{
		AssignmentID: assigned.Assignment.ID,
		ReleasedOn:   times.Date(2026, 9, 30),
	})
	require.NoError(t, err)

	// act
	_, err = sut.Do(t.Context(), usecase.ReleaseAssignmentInput{
		AssignmentID: assigned.Assignment.ID,
		ReleasedOn:   times.Date(2026, 10, 31),
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrPrecondition)
}

func TestReleaseAssignment_Do_PrecedesAssignment(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	employeeID := hireEmployee(t, infra)
	departmentID := createDepartment(t, infra)
	assigned, err := usecase.NewAssignEmployee(infra).Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2026, 4, 1),
	})
	require.NoError(t, err)
	sut := usecase.NewReleaseAssignment(infra)

	// act
	_, err = sut.Do(t.Context(), usecase.ReleaseAssignmentInput{
		AssignmentID: assigned.Assignment.ID,
		ReleasedOn:   times.Date(2026, 3, 31),
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrPrecondition)
}
