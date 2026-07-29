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

func TestAssignEmployee_Do(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	employeeID := hireEmployee(t, infra)
	departmentID := createDepartment(t, infra)
	sut := usecase.NewAssignEmployee(infra)

	// act
	out, err := sut.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2026, 4, 1),
	})

	// assert
	require.NoError(t, err)
	assert.NotEmpty(t, out.Assignment.ID)
	assert.Equal(t, employeeID, out.Assignment.EmployeeID)
	assert.Equal(t, departmentID, out.Assignment.DepartmentID)
	assert.Equal(t, model.PositionMember, out.Assignment.Position)
	assert.Nil(t, out.Assignment.ReleasedOn)
}

func TestAssignEmployee_Do_EmployeeNotFound(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	departmentID := createDepartment(t, infra)
	sut := usecase.NewAssignEmployee(infra)

	// act
	_, err := sut.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   uuid.New(),
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2026, 4, 1),
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrNotFound)
}

func TestAssignEmployee_Do_DepartmentNotFound(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	employeeID := hireEmployee(t, infra)
	sut := usecase.NewAssignEmployee(infra)

	// act
	_, err := sut.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: uuid.New(),
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2026, 4, 1),
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrNotFound)
}

func TestAssignEmployee_Do_AlreadyAssigned(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	employeeID := hireEmployee(t, infra)
	departmentID := createDepartment(t, infra)
	sut := usecase.NewAssignEmployee(infra)
	_, err := sut.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2026, 4, 1),
	})
	require.NoError(t, err)

	// act
	_, err = sut.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionManager,
		AssignedOn:   times.Date(2026, 5, 1),
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrPrecondition)
}

func TestAssignEmployee_Do_OverlapsReleased(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	employeeID := hireEmployee(t, infra)
	departmentID := createDepartment(t, infra)
	assign := usecase.NewAssignEmployee(infra)
	release := usecase.NewReleaseAssignment(infra)
	assigned, err := assign.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2026, 4, 1),
	})
	require.NoError(t, err)
	_, err = release.Do(t.Context(), usecase.ReleaseAssignmentInput{
		AssignmentID: assigned.Assignment.ID,
		ReleasedOn:   times.Date(2026, 9, 30),
	})
	require.NoError(t, err)

	// act: starting before the previous release must be rejected, on it is fine
	_, err = assign.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2026, 9, 29),
	})
	require.Error(t, err)
	require.ErrorIs(t, err, aerrors.ErrPrecondition)

	_, err = assign.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2026, 9, 30),
	})

	// assert
	require.NoError(t, err)
}
