package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/feature/assignment/model"
	"github.com/mickamy/employee-management/internal/feature/assignment/repository"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
)

type ListAssignmentHistoryInput struct {
	EmployeeID uuid.UUID
}

type ListAssignmentHistoryOutput struct {
	Assignments []model.Assignment
}

type ListAssignmentHistory struct {
	_                    di.Infra              `inject:"embed"`
	tx                   tx.ReadTransactor     `inject:""`
	assignmentRepository repository.Assignment `inject:""`
}

func (uc ListAssignmentHistory) Do(
	ctx context.Context,
	input ListAssignmentHistoryInput,
) (ListAssignmentHistoryOutput, error) {
	var assignments []model.Assignment
	if err := uc.tx.WithReadTx(ctx, func(tx tx.Tx) error {
		found, err := uc.assignmentRepository.Bind(tx).History(ctx, input.EmployeeID)
		if err != nil {
			return fmt.Errorf("fetch assignment history: %w", err)
		}
		assignments = found
		return nil
	}); err != nil {
		return ListAssignmentHistoryOutput{}, fmt.Errorf("list assignment history: %w", err)
	}

	return ListAssignmentHistoryOutput{Assignments: assignments}, nil
}
