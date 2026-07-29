package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/mickamy/employee-management/internal/errors/aerrors"
	"github.com/mickamy/employee-management/internal/feature/assignment/model"
	"github.com/mickamy/employee-management/internal/feature/assignment/repository"
)

type assignmentStream struct {
	assignment model.Assignment
	stream     model.Stream
}

func findAssignmentStream(
	ctx context.Context,
	boundedRepository repository.Stream,
	id uuid.UUID,
) (assignmentStream, error) {
	streamID, err := boundedRepository.FindByAssignment(ctx, id)
	if err != nil {
		return assignmentStream{}, fmt.Errorf("find stream: %w", err)
	}
	stream, err := boundedRepository.Load(ctx, streamID)
	if err != nil {
		return assignmentStream{}, fmt.Errorf("load stream: %w", err)
	}

	a, ok := stream.Assignment(id)
	if !ok {
		return assignmentStream{}, aerrors.NotFound("assignment")
	}

	return assignmentStream{assignment: a, stream: stream}, nil
}
