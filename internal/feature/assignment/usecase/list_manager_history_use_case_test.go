package usecase_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/lib/times"

	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/feature/assignment/model"
	"github.com/mickamy/employee-management/internal/feature/assignment/usecase"
	efixture "github.com/mickamy/employee-management/internal/feature/employee/fixture"
	eusecase "github.com/mickamy/employee-management/internal/feature/employee/usecase"
	"github.com/mickamy/employee-management/internal/test/tinfra"
)

func hireNamedEmployee(t *testing.T, infra di.Infra, name string) uuid.UUID {
	t.Helper()

	e := efixture.Employee()
	out, err := eusecase.NewHireEmployee(infra).Do(t.Context(), eusecase.HireEmployeeInput{
		Code:    e.Code,
		Name:    name,
		Email:   e.Email,
		HiredOn: e.HiredOn.Truncate(24 * time.Hour),
	})
	require.NoError(t, err)
	return out.Employee.ID
}

func TestListManagerHistory_Do(t *testing.T) {
	t.Parallel()

	// arrange: a member and a manager overlap in the same department
	infra := tinfra.New(t)
	memberID := hireNamedEmployee(t, infra, "Member")
	managerID := hireNamedEmployee(t, infra, "Manager")
	departmentID := createDepartment(t, infra)
	assign := usecase.NewAssignEmployee(infra)
	_, err := assign.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   memberID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2026, 4, 1),
	})
	require.NoError(t, err)
	_, err = assign.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   managerID,
		DepartmentID: departmentID,
		Position:     model.PositionManager,
		AssignedOn:   times.Date(2026, 6, 1),
	})
	require.NoError(t, err)
	sut := usecase.NewListManagerHistory(infra)

	// act
	out, err := sut.Do(t.Context(), usecase.ListManagerHistoryInput{
		EmployeeID: memberID,
	})

	// assert: the tenure starts when the later of the two assignments starts
	require.NoError(t, err)
	require.Len(t, out.Tenures, 1)
	tenure := out.Tenures[0]
	assert.Equal(t, managerID, tenure.ManagerEmployeeID)
	assert.Equal(t, "Manager", tenure.ManagerName)
	assert.Equal(t, departmentID, tenure.DepartmentID)
	assert.Equal(t, times.Date(2026, 6, 1), tenure.StartedOn)
	assert.Nil(t, tenure.EndedOn)
}

func TestListManagerHistory_Do_ManagerChangeSplitsTenures(t *testing.T) {
	t.Parallel()

	// arrange: the department head changes while the member stays
	infra := tinfra.New(t)
	memberID := hireNamedEmployee(t, infra, "Member")
	firstManagerID := hireNamedEmployee(t, infra, "First Manager")
	secondManagerID := hireNamedEmployee(t, infra, "Second Manager")
	departmentID := createDepartment(t, infra)
	assign := usecase.NewAssignEmployee(infra)
	release := usecase.NewReleaseAssignment(infra)

	_, err := assign.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   memberID,
		DepartmentID: departmentID,
		Position:     model.PositionMember,
		AssignedOn:   times.Date(2026, 1, 1),
	})
	require.NoError(t, err)
	firstAssigned, err := assign.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   firstManagerID,
		DepartmentID: departmentID,
		Position:     model.PositionManager,
		AssignedOn:   times.Date(2026, 1, 1),
	})
	require.NoError(t, err)
	_, err = release.Do(t.Context(), usecase.ReleaseAssignmentInput{
		AssignmentID: firstAssigned.Assignment.ID,
		ReleasedOn:   times.Date(2026, 6, 1),
	})
	require.NoError(t, err)
	_, err = assign.Do(t.Context(), usecase.AssignEmployeeInput{
		EmployeeID:   secondManagerID,
		DepartmentID: departmentID,
		Position:     model.PositionManager,
		AssignedOn:   times.Date(2026, 6, 1),
	})
	require.NoError(t, err)
	sut := usecase.NewListManagerHistory(infra)

	// act
	out, err := sut.Do(t.Context(), usecase.ListManagerHistoryInput{
		EmployeeID: memberID,
	})

	// assert: a reorganization shows up in manager history by construction
	require.NoError(t, err)
	require.Len(t, out.Tenures, 2)
	assert.Equal(t, firstManagerID, out.Tenures[0].ManagerEmployeeID)
	assert.Equal(t, "First Manager", out.Tenures[0].ManagerName)
	assert.Equal(t, times.Date(2026, 1, 1), out.Tenures[0].StartedOn)
	require.NotNil(t, out.Tenures[0].EndedOn)
	assert.Equal(t, times.Date(2026, 6, 1), *out.Tenures[0].EndedOn)
	assert.Equal(t, secondManagerID, out.Tenures[1].ManagerEmployeeID)
	assert.Equal(t, times.Date(2026, 6, 1), out.Tenures[1].StartedOn)
	assert.Nil(t, out.Tenures[1].EndedOn)
}
