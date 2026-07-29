package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/errors/aerrors"
	"github.com/mickamy/employee-management/internal/feature/assignment/model"
	"github.com/mickamy/employee-management/internal/feature/assignment/repository"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
	"github.com/mickamy/employee-management/internal/lib/clock"
)

type RevokeReleaseInput struct {
	AssignmentID uuid.UUID
	Reason       string
}

type RevokeReleaseOutput struct{}

type RevokeRelease struct {
	_                          di.Infra                  `inject:"embed"`
	tx                         tx.Transactor             `inject:""`
	assignmentStreamRepository repository.Stream         `inject:""`
	assignmentViewRepository   repository.AssignmentView `inject:""`
}

func (uc RevokeRelease) Do(ctx context.Context, input RevokeReleaseInput) (RevokeReleaseOutput, error) {
	if err := uc.tx.WithTx(ctx, func(tx tx.Tx) error {
		bound := uc.assignmentStreamRepository.Bind(tx)
		as, err := findAssignmentStream(ctx, bound, input.AssignmentID)
		if err != nil {
			return err
		}

		if as.assignment.ReleasedOn == nil {
			return aerrors.Precondition("assignment is not released")
		}
		// Undoing a release already in effect is a retroactive correction,
		// which is a different business activity.
		if !as.assignment.ReleasedOn.After(clock.Now(ctx)) {
			return aerrors.Precondition("release already in effect")
		}
		// Reopening the assignment must not overlap a later one.
		if latest, ok := as.stream.Latest(); ok && latest.ID != as.assignment.ID {
			return aerrors.Precondition("a later assignment exists")
		}

		if _, err := bound.Append(ctx, as.stream, model.ReleaseRevoked{
			AssignmentID: input.AssignmentID,
			Reason:       input.Reason,
		}); err != nil {
			return fmt.Errorf("append release revocation: %w", err)
		}
		if err := uc.assignmentViewRepository.Bind(tx).ClearRelease(ctx, input.AssignmentID); err != nil {
			return fmt.Errorf("project release revocation: %w", err)
		}
		return nil
	}); err != nil {
		return RevokeReleaseOutput{}, fmt.Errorf("revoke release: %w", err)
	}

	return RevokeReleaseOutput{}, nil
}
