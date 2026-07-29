package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/lib/times"

	"github.com/mickamy/employee-management/internal/errors/aerrors"
	"github.com/mickamy/employee-management/internal/feature/assignment/model"
	"github.com/mickamy/employee-management/internal/feature/assignment/usecase"
	"github.com/mickamy/employee-management/internal/test/tinfra"
)

func TestRevokeRelease_Do(t *testing.T) {
	t.Parallel()

	// arrange: a release decided ahead of its effective date
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
	_, err = usecase.NewReleaseAssignment(infra).Do(t.Context(), usecase.ReleaseAssignmentInput{
		AssignmentID: assigned.Assignment.ID,
		ReleasedOn:   times.Date(2100, 9, 30),
	})
	require.NoError(t, err)
	sut := usecase.NewRevokeRelease(infra)

	// act
	_, err = sut.Do(t.Context(), usecase.RevokeReleaseInput{
		AssignmentID: assigned.Assignment.ID,
		Reason:       "plans changed",
	})

	// assert: the assignment is open again
	require.NoError(t, err)
	current, err := usecase.NewGetCurrentAssignment(infra).Do(t.Context(), usecase.GetCurrentAssignmentInput{
		EmployeeID: employeeID,
	})
	require.NoError(t, err)
	assert.Equal(t, assigned.Assignment.ID, current.Assignment.ID)
	assert.Nil(t, current.Assignment.ReleasedOn)
}

func TestRevokeRelease_Do_NotReleased(t *testing.T) {
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
	sut := usecase.NewRevokeRelease(infra)

	// act
	_, err = sut.Do(t.Context(), usecase.RevokeReleaseInput{
		AssignmentID: assigned.Assignment.ID,
		Reason:       "nothing to revoke",
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrPrecondition)
}

func TestRevokeRelease_Do_AlreadyInEffect(t *testing.T) {
	t.Parallel()

	// arrange: the release date has already passed
	infra := tinfra.New(t)
	employeeID := hireEmployee(t, infra)
	departmentID := createDepartment(t, infra)
	assigned, err := usecase.NewAssignEmployee(infra).Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2020, 4, 1),
	})
	require.NoError(t, err)
	_, err = usecase.NewReleaseAssignment(infra).Do(t.Context(), usecase.ReleaseAssignmentInput{
		AssignmentID: assigned.Assignment.ID,
		ReleasedOn:   times.Date(2021, 3, 31),
	})
	require.NoError(t, err)
	sut := usecase.NewRevokeRelease(infra)

	// act
	_, err = sut.Do(t.Context(), usecase.RevokeReleaseInput{
		AssignmentID: assigned.Assignment.ID,
		Reason:       "too late",
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrPrecondition)
}

func TestRevokeRelease_Do_LaterAssignmentExists(t *testing.T) {
	t.Parallel()

	// arrange: a successor assignment was already decided
	infra := tinfra.New(t)
	employeeID := hireEmployee(t, infra)
	departmentID := createDepartment(t, infra)
	assign := usecase.NewAssignEmployee(infra)
	first, err := assign.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2026, 4, 1),
	})
	require.NoError(t, err)
	_, err = usecase.NewReleaseAssignment(infra).Do(t.Context(), usecase.ReleaseAssignmentInput{
		AssignmentID: first.Assignment.ID,
		ReleasedOn:   times.Date(2100, 9, 30),
	})
	require.NoError(t, err)
	_, err = assign.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionManager,
		AssignedOn:   times.Date(2100, 9, 30),
	})
	require.NoError(t, err)
	sut := usecase.NewRevokeRelease(infra)

	// act: reopening the first assignment would overlap the successor
	_, err = sut.Do(t.Context(), usecase.RevokeReleaseInput{
		AssignmentID: first.Assignment.ID,
		Reason:       "plans changed",
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrPrecondition)
}
