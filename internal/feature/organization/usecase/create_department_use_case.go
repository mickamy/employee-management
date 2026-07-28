// Package usecase implements the organization feature's business operations.
package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/feature/organization/model"
	"github.com/mickamy/employee-management/internal/feature/organization/repository"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
)

type CreateDepartmentInput struct {
	Code string
	Name string
}

type CreateDepartmentOutput struct {
	Department model.Department
}

type CreateDepartment struct {
	_                    di.Infra              `inject:"embed"`
	tx                   tx.Transactor         `inject:""`
	departmentRepository repository.Department `inject:""`
}

func (uc CreateDepartment) Do(ctx context.Context, input CreateDepartmentInput) (CreateDepartmentOutput, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return CreateDepartmentOutput{}, fmt.Errorf("new department id: %w", err)
	}

	d := model.Department{
		ID:   id,
		Code: input.Code,
		Name: input.Name,
	}

	if err := uc.tx.WithTx(ctx, func(tx tx.Tx) error {
		if err := uc.departmentRepository.Bind(tx).Create(ctx, d); err != nil {
			return fmt.Errorf("insert department: %w", err)
		}
		return nil
	}); err != nil {
		return CreateDepartmentOutput{}, fmt.Errorf("create department: %w", err)
	}

	return CreateDepartmentOutput{Department: d}, nil
}
