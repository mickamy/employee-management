package usecase_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mickamy/employee-management/internal/feature/employee/fixture"
	"github.com/mickamy/employee-management/internal/feature/employee/model"
	"github.com/mickamy/employee-management/internal/feature/employee/repository"
	"github.com/mickamy/employee-management/internal/feature/employee/usecase"
	"github.com/mickamy/employee-management/internal/storage/tx"
	"github.com/mickamy/employee-management/internal/test/tinfra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListEmployee_Do(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	employee1, hire1 := fixture.EmployeeAndHire()
	employee2, hire2 := fixture.EmployeeAndHire()
	for _, m := range []struct {
		employee model.Employee
		hire     model.EmployeeHire
	}{
		{
			employee: employee1,
			hire:     hire1,
		},
		{
			employee: employee2,
			hire:     hire2,
		},
	} {
		require.NoError(t, tx.NewTransactor(infra.Writer).WithTx(t.Context(), func(tx tx.Tx) error {
			bound := repository.NewEmployee(infra.Reader).Bind(tx)
			if err := bound.Create(t.Context(), m.employee); err != nil {
				return err
			}
			return bound.CreateHire(t.Context(), m.hire)
		}))
	}
	sut := usecase.NewListEmployees(infra)

	// act
	out, err := sut.Do(t.Context(), usecase.ListEmployeesInput{
		AfterID:  uuid.Nil,
		PageSize: 10,
	})

	// assert
	require.NoError(t, err)
	assert.Equal(t, employee1, out.Employees[0])
	assert.Equal(t, employee2, out.Employees[1])
}
