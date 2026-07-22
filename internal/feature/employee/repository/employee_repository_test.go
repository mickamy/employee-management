package repository_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/feature/employee/model"
	"github.com/mickamy/employee-management/internal/feature/employee/repository"
	"github.com/mickamy/employee-management/internal/test/tdb"
)

func TestUnboundWritesAreRejected(t *testing.T) {
	t.Parallel()

	d := tdb.New(t)
	repo := repository.NewEmployee(d.Reader)

	err := repo.Create(t.Context(), model.Employee{
		ID:    uuid.New(),
		Code:  "E0001",
		Name:  "Test Employee",
		Email: "test@example.com",
	})
	require.ErrorContains(t, err, "permission denied")
}
