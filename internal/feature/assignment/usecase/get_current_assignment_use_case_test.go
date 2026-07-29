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

func TestGetCurrentAssignment_Do_ReadsOwnWrite(t *testing.T) {
	t.Parallel()

	// arrange: the projection is transactional, so the write is immediately visible
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
	sut := usecase.NewGetCurrentAssignment(infra)

	// act
	out, err := sut.Do(t.Context(), usecase.GetCurrentAssignmentInput{
		EmployeeID: employeeID,
	})

	// assert
	require.NoError(t, err)
	assert.Equal(t, assigned.Assignment.ID, out.Assignment.ID)
	assert.Equal(t, departmentID, out.Assignment.DepartmentID)
}

func TestGetCurrentAssignment_Do_FutureDatedIsNotCurrent(t *testing.T) {
	t.Parallel()

	// arrange: the decision exists but only takes effect in the future
	infra := tinfra.New(t)
	employeeID := hireEmployee(t, infra)
	departmentID := createDepartment(t, infra)
	_, err := usecase.NewAssignEmployee(infra).Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2100, 4, 1),
	})
	require.NoError(t, err)
	sut := usecase.NewGetCurrentAssignment(infra)

	// act
	_, err = sut.Do(t.Context(), usecase.GetCurrentAssignmentInput{
		EmployeeID: employeeID,
	})

	// assert: nothing is current today, yet the history shows the decision
	require.Error(t, err)
	require.ErrorIs(t, err, aerrors.ErrNotFound)
	history, err := usecase.NewListAssignmentHistory(infra).Do(t.Context(), usecase.ListAssignmentHistoryInput{
		EmployeeID: employeeID,
	})
	require.NoError(t, err)
	assert.Len(t, history.Assignments, 1)
}
