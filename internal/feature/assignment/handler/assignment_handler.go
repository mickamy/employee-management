package handler

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	assignmentv1 "github.com/mickamy/employee-management/gen/assignment/v1"
	"github.com/mickamy/employee-management/gen/assignment/v1/assignmentv1connect"
	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/errors/cerrors"
	"github.com/mickamy/employee-management/internal/feature/assignment/converter"
	"github.com/mickamy/employee-management/internal/feature/assignment/usecase"
	"github.com/mickamy/employee-management/internal/lib/converters"
)

// Assignment serves both the command and the query service: the CQRS split
// is a contract property, not a deployment one.
type Assignment struct {
	_             di.Infra                       `inject:"embed"`
	assign        *usecase.AssignEmployee        `inject:""`
	release       *usecase.ReleaseAssignment     `inject:""`
	revoke        *usecase.RevokeAssignment      `inject:""`
	revokeRelease *usecase.RevokeRelease         `inject:""`
	current       *usecase.GetCurrentAssignment  `inject:""`
	history       *usecase.ListAssignmentHistory `inject:""`
	manager       *usecase.ListManagerHistory    `inject:""`
}

var (
	_ assignmentv1connect.AssignmentCommandServiceHandler = (*Assignment)(nil)
	_ assignmentv1connect.AssignmentQueryServiceHandler   = (*Assignment)(nil)
)

func (h *Assignment) AssignEmployee(
	ctx context.Context,
	req *connect.Request[assignmentv1.AssignEmployeeRequest],
) (*connect.Response[assignmentv1.AssignEmployeeResponse], error) {
	employeeID, err := uuid.Parse(req.Msg.GetEmployeeId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	departmentID, err := uuid.Parse(req.Msg.GetDepartmentId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.assign.Do(ctx, usecase.AssignEmployeeInput{
		EmployeeID:   employeeID,
		DepartmentID: departmentID,
		Position:     converters.ToAssignmentPosition(req.Msg.GetPosition()),
		AssignedOn:   converters.ToTime(req.Msg.GetAssignedOn()),
	})
	if err != nil {
		return nil, cerrors.Map(err)
	}

	return connect.NewResponse(&assignmentv1.AssignEmployeeResponse{
		Assignment: converter.AssignmentToAssignmentv1(out.Assignment),
	}), nil
}

func (h *Assignment) ReleaseAssignment(
	ctx context.Context,
	req *connect.Request[assignmentv1.ReleaseAssignmentRequest],
) (*connect.Response[assignmentv1.ReleaseAssignmentResponse], error) {
	assignmentID, err := uuid.Parse(req.Msg.GetAssignmentId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.release.Do(ctx, usecase.ReleaseAssignmentInput{
		AssignmentID: assignmentID,
		ReleasedOn:   converters.ToTime(req.Msg.GetReleasedOn()),
	})
	if err != nil {
		return nil, cerrors.Map(err)
	}

	return connect.NewResponse(&assignmentv1.ReleaseAssignmentResponse{
		Assignment: converter.AssignmentToAssignmentv1(out.Assignment),
	}), nil
}

func (h *Assignment) RevokeAssignment(
	ctx context.Context,
	req *connect.Request[assignmentv1.RevokeAssignmentRequest],
) (*connect.Response[assignmentv1.RevokeAssignmentResponse], error) {
	assignmentID, err := uuid.Parse(req.Msg.GetAssignmentId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if _, err := h.revoke.Do(ctx, usecase.RevokeAssignmentInput{
		AssignmentID: assignmentID,
		Reason:       req.Msg.GetReason(),
	}); err != nil {
		return nil, cerrors.Map(err)
	}

	return connect.NewResponse(&assignmentv1.RevokeAssignmentResponse{}), nil
}

func (h *Assignment) RevokeRelease(
	ctx context.Context,
	req *connect.Request[assignmentv1.RevokeReleaseRequest],
) (*connect.Response[assignmentv1.RevokeReleaseResponse], error) {
	assignmentID, err := uuid.Parse(req.Msg.GetAssignmentId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if _, err := h.revokeRelease.Do(ctx, usecase.RevokeReleaseInput{
		AssignmentID: assignmentID,
		Reason:       req.Msg.GetReason(),
	}); err != nil {
		return nil, cerrors.Map(err)
	}

	return connect.NewResponse(&assignmentv1.RevokeReleaseResponse{}), nil
}

func (h *Assignment) GetCurrentAssignment(
	ctx context.Context,
	req *connect.Request[assignmentv1.GetCurrentAssignmentRequest],
) (*connect.Response[assignmentv1.GetCurrentAssignmentResponse], error) {
	employeeID, err := uuid.Parse(req.Msg.GetEmployeeId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.current.Do(ctx, usecase.GetCurrentAssignmentInput{
		EmployeeID: employeeID,
	})
	if err != nil {
		return nil, cerrors.Map(err)
	}

	return connect.NewResponse(&assignmentv1.GetCurrentAssignmentResponse{
		Assignment: converter.AssignmentToAssignmentv1(out.Assignment),
	}), nil
}

func (h *Assignment) ListAssignmentHistory(
	ctx context.Context,
	req *connect.Request[assignmentv1.ListAssignmentHistoryRequest],
) (*connect.Response[assignmentv1.ListAssignmentHistoryResponse], error) {
	employeeID, err := uuid.Parse(req.Msg.GetEmployeeId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.history.Do(ctx, usecase.ListAssignmentHistoryInput{
		EmployeeID: employeeID,
	})
	if err != nil {
		return nil, cerrors.Map(err)
	}

	items := make([]*assignmentv1.Assignment, len(out.Assignments))
	for i := range out.Assignments {
		items[i] = converter.AssignmentToAssignmentv1(out.Assignments[i])
	}
	return connect.NewResponse(&assignmentv1.ListAssignmentHistoryResponse{
		Assignments: items,
	}), nil
}

func (h *Assignment) ListManagerHistory(
	ctx context.Context,
	req *connect.Request[assignmentv1.ListManagerHistoryRequest],
) (*connect.Response[assignmentv1.ListManagerHistoryResponse], error) {
	employeeID, err := uuid.Parse(req.Msg.GetEmployeeId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.manager.Do(ctx, usecase.ListManagerHistoryInput{
		EmployeeID: employeeID,
	})
	if err != nil {
		return nil, cerrors.Map(err)
	}

	items := make([]*assignmentv1.ManagerTenure, len(out.Tenures))
	for i := range out.Tenures {
		items[i] = converter.ManagerTenureToAssignmentv1(out.Tenures[i])
	}
	return connect.NewResponse(&assignmentv1.ListManagerHistoryResponse{
		Tenures: items,
	}), nil
}
