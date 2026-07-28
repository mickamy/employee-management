package usecase

import (
	"context"
	"fmt"

	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/feature/organization/model"
	"github.com/mickamy/employee-management/internal/feature/organization/repository"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
)

type ListDepartmentsInput struct{}

type ListDepartmentsOutput struct {
	Departments []model.Department
}

type ListDepartments struct {
	_                    di.Infra              `inject:"embed"`
	tx                   tx.ReadTransactor     `inject:""`
	departmentRepository repository.Department `inject:""`
}

func (uc ListDepartments) Do(ctx context.Context, _ ListDepartmentsInput) (ListDepartmentsOutput, error) {
	var departments []model.Department
	if err := uc.tx.WithReadTx(ctx, func(tx tx.Tx) error {
		found, err := uc.departmentRepository.Bind(tx).List(ctx)
		if err != nil {
			return fmt.Errorf("fetch departments: %w", err)
		}
		departments = found
		return nil
	}); err != nil {
		return ListDepartmentsOutput{}, fmt.Errorf("list departments: %w", err)
	}

	return ListDepartmentsOutput{Departments: departments}, nil
}
