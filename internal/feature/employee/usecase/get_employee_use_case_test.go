package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/feature/employee/fixture"
	"github.com/mickamy/employee-management/internal/feature/employee/repository"
	"github.com/mickamy/employee-management/internal/feature/employee/usecase"
	"github.com/mickamy/employee-management/internal/storage/tx"
	"github.com/mickamy/employee-management/internal/test/tinfra"
)

func TestGetEmployee_Do(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	sut := usecase.NewGetEmployee(infra)
	employee, hire := fixture.EmployeeAndHire()
	require.NoError(t, tx.NewTransactor(infra.Writer).WithTx(t.Context(), func(tx tx.Tx) error {
		bound := repository.NewEmployee(infra.Reader).Bind(tx)
		if err := bound.Create(t.Context(), employee); err != nil {
			return err
		}
		return bound.CreateHire(t.Context(), hire)
	}))

	// act
	out, err := sut.Do(t.Context(), usecase.GetEmployeeInput{
		ID: employee.ID,
	})

	// assert
	require.NoError(t, err)
	assert.Equal(t, employee, out.Employee)
}
