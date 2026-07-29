package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/errors/aerrors"
	"github.com/mickamy/employee-management/internal/feature/assignment/model"
	"github.com/mickamy/employee-management/internal/feature/assignment/repository"
	erepository "github.com/mickamy/employee-management/internal/feature/employee/repository"
	orepository "github.com/mickamy/employee-management/internal/feature/organization/repository"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
)

type AssignEmployeeInput struct {
	EmployeeID   uuid.UUID
	DepartmentID uuid.UUID
	Position     model.Position
	AssignedOn   time.Time
}

type AssignEmployeeOutput struct {
	Assignment model.Assignment
}

type AssignEmployee struct {
	_                          di.Infra                  `inject:"embed"`
	tx                         tx.Transactor             `inject:""`
	assignmentStreamRepository repository.Stream         `inject:""`
	assignmentViewRepository   repository.AssignmentView `inject:""`
	employeeRepository         erepository.Employee      `inject:""`
	departmentRepository       orepository.Department    `inject:""`
}

func (uc AssignEmployee) Do(ctx context.Context, input AssignEmployeeInput) (AssignEmployeeOutput, error) {
	assignmentID, err := uuid.NewV7()
	if err != nil {
		return AssignEmployeeOutput{}, fmt.Errorf("new assignment id: %w", err)
	}

	a := model.Assignment{
		ID:           assignmentID,
		EmployeeID:   input.EmployeeID,
		DepartmentID: input.DepartmentID,
		Position:     input.Position,
		AssignedOn:   input.AssignedOn,
	}

	if err := uc.tx.WithTx(ctx, func(tx tx.Tx) error {
		_, err := uc.employeeRepository.Bind(tx).Find(ctx, input.EmployeeID)
		if err != nil {
			if errors.Is(err, aerrors.ErrNotFound) {
				return aerrors.NotFound("employee")
			}
			return fmt.Errorf("find employee: %w", err)
		}

		exists, err := uc.departmentRepository.Bind(tx).Exists(ctx, input.DepartmentID)
		if err != nil {
			return fmt.Errorf("check department: %w", err)
		}
		if !exists {
			return aerrors.NotFound("department")
		}

		bound := uc.assignmentStreamRepository.Bind(tx)
		stream, err := bound.Load(ctx, input.EmployeeID)
		if err != nil {
			return fmt.Errorf("load stream: %w", err)
		}
		if _, open := stream.Open(); open {
			return aerrors.Precondition("employee already has an assignment")
		}
		if last, released := stream.LastReleasedOn(); released && input.AssignedOn.Before(last) {
			return aerrors.Precondition("assignment overlaps a released assignment")
		}

		if _, err := bound.Append(ctx, stream, model.AssignmentDecided{
			AssignmentID: assignmentID,
			DepartmentID: input.DepartmentID,
			Position:     input.Position,
			AssignedOn:   input.AssignedOn,
		}); err != nil {
			return fmt.Errorf("append decision: %w", err)
		}
		if err := uc.assignmentViewRepository.Bind(tx).Insert(ctx, a); err != nil {
			return fmt.Errorf("project decision: %w", err)
		}
		return nil
	}); err != nil {
		return AssignEmployeeOutput{}, fmt.Errorf("assign employee: %w", err)
	}

	return AssignEmployeeOutput{Assignment: a}, nil
}
