package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/feature/assignment/model"
	"github.com/mickamy/employee-management/internal/feature/assignment/repository"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
	"github.com/mickamy/employee-management/internal/lib/clock"
)

type GetCurrentAssignmentInput struct {
	EmployeeID uuid.UUID
}

type GetCurrentAssignmentOutput struct {
	Assignment model.Assignment
}

type GetCurrentAssignment struct {
	_                    di.Infra              `inject:"embed"`
	tx                   tx.ReadTransactor     `inject:""`
	assignmentRepository repository.Assignment `inject:""`
}

func (uc GetCurrentAssignment) Do(
	ctx context.Context,
	input GetCurrentAssignmentInput,
) (GetCurrentAssignmentOutput, error) {
	var assignment model.Assignment
	if err := uc.tx.WithReadTx(ctx, func(tx tx.Tx) error {
		found, err := uc.assignmentRepository.Bind(tx).Current(ctx, input.EmployeeID, clock.Now(ctx))
		if err != nil {
			return fmt.Errorf("fetch current assignment: %w", err)
		}
		assignment = found
		return nil
	}); err != nil {
		return GetCurrentAssignmentOutput{}, fmt.Errorf("get current assignment: %w", err)
	}

	return GetCurrentAssignmentOutput{Assignment: assignment}, nil
}
