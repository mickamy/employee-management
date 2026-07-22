// Package usecase implements the employee feature's business operations.
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mickamy/employee-management/internal/feature/employee/model"
	"github.com/mickamy/employee-management/internal/feature/employee/repository"
	"github.com/mickamy/employee-management/internal/storage/tx"
)

type HireEmployeeInput struct {
	Code    string
	Name    string
	Email   string
	HiredOn time.Time
}

type HireEmployeeOutput struct {
	Employee model.Employee
}

type HireEmployee struct {
	tx                 tx.Transactor       `inject:"arg"`
	employeeRepository repository.Employee `inject:"arg"`
}

func (uc HireEmployee) Do(ctx context.Context, input HireEmployeeInput) (HireEmployeeOutput, error) {
	employeeID, err := uuid.NewV7()
	if err != nil {
		return HireEmployeeOutput{}, fmt.Errorf("new employee id: %w", err)
	}
	hireEventID, err := uuid.NewV7()
	if err != nil {
		return HireEmployeeOutput{}, fmt.Errorf("new hire event id: %w", err)
	}

	e := model.Employee{
		ID:      employeeID,
		Code:    input.Code,
		Name:    input.Name,
		Email:   input.Email,
		HiredOn: input.HiredOn,
	}
	hire := model.EmployeeHire{
		ID:         hireEventID,
		EmployeeID: employeeID,
		HiredOn:    input.HiredOn,
	}

	if err = uc.tx.WithTx(ctx, func(tx tx.Tx) error {
		etx := uc.employeeRepository.Bind(tx)
		if err := etx.Create(ctx, e); err != nil {
			return fmt.Errorf("create employee: %w", err)
		}
		if err := etx.CreateHire(ctx, hire); err != nil {
			return fmt.Errorf("create hire event: %w", err)
		}
		return nil
	}); err != nil {
		return HireEmployeeOutput{}, fmt.Errorf("hire employee: %w", err)
	}

	return HireEmployeeOutput{Employee: e}, nil
}
