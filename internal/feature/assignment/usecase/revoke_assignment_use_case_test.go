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

func TestRevokeAssignment_Do(t *testing.T) {
	t.Parallel()

	// arrange: a decision that has not yet taken effect
	infra := tinfra.New(t)
	ctx := fixedNow(t)
	employeeID := hireEmployee(t, infra)
	departmentID := createDepartment(t, infra)
	assigned, err := usecase.NewAssignEmployee(infra).Do(ctx, usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2100, 4, 1),
	})
	require.NoError(t, err)
	sut := usecase.NewRevokeAssignment(infra)

	// act
	_, err = sut.Do(ctx, usecase.RevokeAssignmentInput{
		AssignmentID: assigned.Assignment.ID,
		Reason:       "wrong department",
	})

	// assert: the revoked decision no longer appears in the clean view
	require.NoError(t, err)
	history, err := usecase.NewListAssignmentHistory(infra).Do(ctx, usecase.ListAssignmentHistoryInput{
		EmployeeID: employeeID,
	})
	require.NoError(t, err)
	assert.Empty(t, history.Assignments)
}

func TestRevokeAssignment_Do_AlreadyInEffect(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	ctx := fixedNow(t)
	employeeID := hireEmployee(t, infra)
	departmentID := createDepartment(t, infra)
	assigned, err := usecase.NewAssignEmployee(infra).Do(ctx, usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2026, 4, 1),
	})
	require.NoError(t, err)
	sut := usecase.NewRevokeAssignment(infra)

	// act
	_, err = sut.Do(ctx, usecase.RevokeAssignmentInput{
		AssignmentID: assigned.Assignment.ID,
		Reason:       "too late",
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrPrecondition)
}

func TestRevokeAssignment_Do_ReleasedAssignment(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	ctx := fixedNow(t)
	employeeID := hireEmployee(t, infra)
	departmentID := createDepartment(t, infra)
	assigned, err := usecase.NewAssignEmployee(infra).Do(ctx, usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2100, 4, 1),
	})
	require.NoError(t, err)
	_, err = usecase.NewReleaseAssignment(infra).Do(ctx, usecase.ReleaseAssignmentInput{
		AssignmentID: assigned.Assignment.ID,
		ReleasedOn:   times.Date(2100, 9, 30),
	})
	require.NoError(t, err)
	sut := usecase.NewRevokeAssignment(infra)

	// act
	_, err = sut.Do(ctx, usecase.RevokeAssignmentInput{
		AssignmentID: assigned.Assignment.ID,
		Reason:       "changed plans",
	})

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, aerrors.ErrPrecondition)
}
