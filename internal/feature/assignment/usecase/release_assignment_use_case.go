package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/errors/aerrors"
	"github.com/mickamy/employee-management/internal/feature/assignment/model"
	"github.com/mickamy/employee-management/internal/feature/assignment/repository"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
)

type ReleaseAssignmentInput struct {
	AssignmentID uuid.UUID
	ReleasedOn   time.Time
}

type ReleaseAssignmentOutput struct {
	Assignment model.Assignment
}

type ReleaseAssignment struct {
	_                          di.Infra                  `inject:"embed"`
	tx                         tx.Transactor             `inject:""`
	assignmentStreamRepository repository.Stream         `inject:""`
	assignmentViewRepository   repository.AssignmentView `inject:""`
}

func (uc ReleaseAssignment) Do(ctx context.Context, input ReleaseAssignmentInput) (ReleaseAssignmentOutput, error) {
	var released model.Assignment
	if err := uc.tx.WithTx(ctx, func(tx tx.Tx) error {
		bound := uc.assignmentStreamRepository.Bind(tx)
		as, err := findAssignmentStream(ctx, bound, input.AssignmentID)
		if err != nil {
			return err
		}

		if as.assignment.ReleasedOn != nil {
			return aerrors.Precondition("assignment already released")
		}
		if input.ReleasedOn.Before(as.assignment.AssignedOn) {
			return aerrors.Precondition("release precedes the assignment")
		}

		if _, err := bound.Append(ctx, as.stream, model.ReleaseDecided{
			AssignmentID: input.AssignmentID,
			ReleasedOn:   input.ReleasedOn,
		}); err != nil {
			return fmt.Errorf("append release: %w", err)
		}
		if err := uc.assignmentViewRepository.Bind(tx).Release(ctx, input.AssignmentID, input.ReleasedOn); err != nil {
			return fmt.Errorf("project release: %w", err)
		}

		as.assignment.ReleasedOn = new(input.ReleasedOn)
		released = as.assignment
		return nil
	}); err != nil {
		return ReleaseAssignmentOutput{}, fmt.Errorf("release assignment: %w", err)
	}

	return ReleaseAssignmentOutput{Assignment: released}, nil
}
