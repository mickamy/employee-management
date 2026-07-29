package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	erepository "github.com/mickamy/employee-management/internal/feature/employee/repository"

	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/feature/assignment/model"
	"github.com/mickamy/employee-management/internal/feature/assignment/repository"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
)

type ListManagerHistoryInput struct {
	EmployeeID uuid.UUID
}

type ListManagerHistoryOutput struct {
	Tenures []model.ManagerTenure
}

type ListManagerHistory struct {
	_                    di.Infra              `inject:"embed"`
	tx                   tx.ReadTransactor     `inject:""`
	assignmentRepository repository.Assignment `inject:""`
	employeeRepository   erepository.Employee  `inject:""`
}

func (uc ListManagerHistory) Do(ctx context.Context, input ListManagerHistoryInput) (ListManagerHistoryOutput, error) {
	var tenures []model.ManagerTenure
	var names map[uuid.UUID]string
	if err := uc.tx.WithReadTx(ctx, func(tx tx.Tx) error {
		found, err := uc.assignmentRepository.Bind(tx).ManagerTenures(ctx, input.EmployeeID)
		if err != nil {
			return fmt.Errorf("fetch manager tenures: %w", err)
		}
		tenures = found
		if len(tenures) == 0 {
			return nil
		}

		ids := make([]uuid.UUID, 0, len(tenures))
		seen := make(map[uuid.UUID]bool, len(tenures))
		for _, t := range tenures {
			if !seen[t.ManagerEmployeeID] {
				seen[t.ManagerEmployeeID] = true
				ids = append(ids, t.ManagerEmployeeID)
			}
		}

		names, err = uc.employeeRepository.Bind(tx).Names(ctx, ids)
		if err != nil {
			return fmt.Errorf("get employee names: %w", err)
		}
		return nil
	}); err != nil {
		return ListManagerHistoryOutput{}, fmt.Errorf("list manager history: %w", err)
	}

	if len(tenures) == 0 {
		return ListManagerHistoryOutput{}, nil
	}

	for i := range tenures {
		tenures[i].ManagerName = names[tenures[i].ManagerEmployeeID]
	}

	return ListManagerHistoryOutput{Tenures: tenures}, nil
}
