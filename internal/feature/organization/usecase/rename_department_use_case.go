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

type RenameDepartmentInput struct {
	ID   uuid.UUID
	Name string
}

type RenameDepartmentOutput struct {
	Department model.Department
}

type RenameDepartment struct {
	_                    di.Infra              `inject:"embed"`
	tx                   tx.Transactor         `inject:""`
	departmentRepository repository.Department `inject:""`
}

func (uc RenameDepartment) Do(ctx context.Context, input RenameDepartmentInput) (RenameDepartmentOutput, error) {
	var d model.Department
	if err := uc.tx.WithTx(ctx, func(tx tx.Tx) error {
		renamed, err := uc.departmentRepository.Bind(tx).Rename(ctx, input.ID, input.Name)
		if err != nil {
			return fmt.Errorf("update department: %w", err)
		}
		d = renamed
		return nil
	}); err != nil {
		return RenameDepartmentOutput{}, fmt.Errorf("rename department: %w", err)
	}

	return RenameDepartmentOutput{Department: d}, nil
}
