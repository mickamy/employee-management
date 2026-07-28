package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/mickamy/employee-management/internal/di"

	"github.com/mickamy/employee-management/internal/feature/employee/model"
	"github.com/mickamy/employee-management/internal/feature/employee/repository"
	"github.com/mickamy/employee-management/internal/storage/tx"
)

type GetEmployeeInput struct {
	ID uuid.UUID
}

type GetEmployeeOutput struct {
	Employee model.Employee
}

type GetEmployee struct {
	_                  di.Infra            `inject:"embed"`
	tx                 tx.ReadTransactor   `inject:""`
	employeeRepository repository.Employee `inject:""`
}

func (uc GetEmployee) Do(ctx context.Context, input GetEmployeeInput) (GetEmployeeOutput, error) {
	var employee model.Employee
	err := uc.tx.WithReadTx(ctx, func(tx tx.Tx) error {
		found, err := uc.employeeRepository.Bind(tx).Find(ctx, input.ID)
		if err != nil {
			return fmt.Errorf("find employee: %w", err)
		}
		employee = found
		return nil
	})
	if err != nil {
		return GetEmployeeOutput{}, fmt.Errorf("get employee: %w", err)
	}

	return GetEmployeeOutput{Employee: employee}, nil
}
