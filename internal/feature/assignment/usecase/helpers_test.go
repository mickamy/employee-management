package usecase_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/di"
	efixture "github.com/mickamy/employee-management/internal/feature/employee/fixture"
	eusecase "github.com/mickamy/employee-management/internal/feature/employee/usecase"
	ofixture "github.com/mickamy/employee-management/internal/feature/organization/fixture"
	ousecase "github.com/mickamy/employee-management/internal/feature/organization/usecase"
)

func hireEmployee(t *testing.T, infra di.Infra) uuid.UUID {
	t.Helper()

	e := efixture.Employee()
	out, err := eusecase.NewHireEmployee(infra).Do(t.Context(), eusecase.HireEmployeeInput{
		Code:    e.Code,
		Name:    e.Name,
		Email:   e.Email,
		HiredOn: e.HiredOn.Truncate(24 * time.Hour),
	})
	require.NoError(t, err)
	return out.Employee.ID
}

func createDepartment(t *testing.T, infra di.Infra) uuid.UUID {
	t.Helper()

	d := ofixture.Department()
	out, err := ousecase.NewCreateDepartment(infra).Do(t.Context(), ousecase.CreateDepartmentInput{
		Code: d.Code,
		Name: d.Name,
	})
	require.NoError(t, err)
	return out.Department.ID
}
