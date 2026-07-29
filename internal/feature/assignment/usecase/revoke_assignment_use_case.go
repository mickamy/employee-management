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

type RevokeAssignmentInput struct {
	AssignmentID uuid.UUID
	Reason       string
}

type RevokeAssignmentOutput struct{}

type RevokeAssignment struct {
	_                          di.Infra                  `inject:"embed"`
	tx                         tx.Transactor             `inject:""`
	assignmentStreamRepository repository.Stream         `inject:""`
	assignmentViewRepository   repository.AssignmentView `inject:""`
}

func (uc RevokeAssignment) Do(ctx context.Context, input RevokeAssignmentInput) (RevokeAssignmentOutput, error) {
	if err := uc.tx.WithTx(ctx, func(tx tx.Tx) error {
		bound := uc.assignmentStreamRepository.Bind(tx)
		as, err := findAssignmentStream(ctx, bound, input.AssignmentID)
		if err != nil {
			return err
		}

		if as.assignment.ReleasedOn != nil {
			return aerrors.Precondition("released assignment cannot be revoked")
		}
		// Undoing a decision already in effect is a release, not a revocation.
		if !as.assignment.AssignedOn.After(clock.Now(ctx)) {
			return aerrors.Precondition("assignment already in effect")
		}

		if _, err := bound.Append(ctx, as.stream, model.AssignmentRevoked{
			AssignmentID: input.AssignmentID,
			Reason:       input.Reason,
		}); err != nil {
			return fmt.Errorf("append revocation: %w", err)
		}
		// The clean view drops revoked decisions; the events stay untouched.
		if err := uc.assignmentViewRepository.Bind(tx).Delete(ctx, input.AssignmentID); err != nil {
			return fmt.Errorf("project revocation: %w", err)
		}
		return nil
	}); err != nil {
		return RevokeAssignmentOutput{}, fmt.Errorf("revoke assignment: %w", err)
	}

	return RevokeAssignmentOutput{}, nil
}
