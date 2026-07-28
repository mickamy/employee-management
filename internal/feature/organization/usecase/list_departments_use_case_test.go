package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/feature/organization/fixture"
	"github.com/mickamy/employee-management/internal/feature/organization/model"
	"github.com/mickamy/employee-management/internal/feature/organization/repository"
	"github.com/mickamy/employee-management/internal/feature/organization/usecase"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
	"github.com/mickamy/employee-management/internal/test/tinfra"
)

func TestListDepartments_Do(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	// created in the opposite order of code, so the returned order can only
	// come from ORDER BY code.
	second := fixture.Department(func(m *model.Department) { m.Code = "b-team" })
	first := fixture.Department(func(m *model.Department) { m.Code = "a-team" })
	require.NoError(t, infra.Transactor.WithTx(t.Context(), func(tx tx.Tx) error {
		bound := repository.NewDepartment(infra.Reader).Bind(tx)
		if err := bound.Create(t.Context(), second); err != nil {
			return err
		}
		return bound.Create(t.Context(), first)
	}))
	sut := usecase.NewListDepartments(infra)

	// act
	out, err := sut.Do(t.Context(), usecase.ListDepartmentsInput{})

	// assert
	require.NoError(t, err)
	require.Len(t, out.Departments, 2)
	assert.Equal(t, first, out.Departments[0])
	assert.Equal(t, second, out.Departments[1])
}
