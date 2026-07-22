package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/mickamy/employee-management/internal/storage/tx"

	"github.com/mickamy/employee-management/internal/feature/employee/model"
	"github.com/mickamy/employee-management/internal/feature/employee/repository"
)

type ListEmployeesInput struct {
	AfterID  uuid.UUID
	PageSize int32
}

type ListEmployeesOutput struct {
	Employees []model.Employee
}

type ListEmployees struct {
	tx                 tx.ReadTransactor   `inject:"arg"`
	employeeRepository repository.Employee `inject:"arg"`
}

func (uc ListEmployees) Do(ctx context.Context, input ListEmployeesInput) (ListEmployeesOutput, error) {
	var employees []model.Employee
	if err := uc.tx.WithReadTx(ctx, func(boundTx tx.Tx) error {
		found, err := uc.employeeRepository.Bind(boundTx).List(ctx, input.AfterID, input.PageSize)
		if err != nil {
			return fmt.Errorf("list employees: %w", err)
		}
		employees = found
		return nil
	}); err != nil {
		return ListEmployeesOutput{}, fmt.Errorf("list employees: %w", err)
	}
	return ListEmployeesOutput{Employees: employees}, nil
}
