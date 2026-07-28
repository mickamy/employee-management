package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/feature/employee/fixture"
	"github.com/mickamy/employee-management/internal/feature/employee/usecase"
	"github.com/mickamy/employee-management/internal/test/tinfra"
)

func TestHireEmployee_Do(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	sut := usecase.NewHireEmployee(infra)
	employee := fixture.Employee()

	// act
	out, err := sut.Do(t.Context(), usecase.HireEmployeeInput{
		Code:    employee.Code,
		Name:    employee.Name,
		Email:   employee.Email,
		HiredOn: employee.HiredOn,
	})

	// assert
	require.NoError(t, err)
	assert.Equal(t, employee.Code, out.Employee.Code)
	assert.Equal(t, employee.Name, out.Employee.Name)
	assert.Equal(t, employee.Email, out.Employee.Email)
	assert.Equal(t, employee.HiredOn, out.Employee.HiredOn)
}
