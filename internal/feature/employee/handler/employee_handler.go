package handler

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/errors/cerrors"

	employeev1 "github.com/mickamy/employee-management/gen/employee/v1"
	"github.com/mickamy/employee-management/gen/employee/v1/employeev1connect"
	"github.com/mickamy/employee-management/internal/feature/employee/converter"
	"github.com/mickamy/employee-management/internal/feature/employee/usecase"
	"github.com/mickamy/employee-management/internal/lib/converters"
)

const defaultPageSize = 50

type Handler struct {
	_    di.Infra               `inject:"embed"`
	hire *usecase.HireEmployee  `inject:""`
	get  *usecase.GetEmployee   `inject:""`
	list *usecase.ListEmployees `inject:""`
}

var _ employeev1connect.EmployeeServiceHandler = (*Handler)(nil)

func (h *Handler) HireEmployee(
	ctx context.Context,
	req *connect.Request[employeev1.HireEmployeeRequest],
) (*connect.Response[employeev1.HireEmployeeResponse], error) {
	out, err := h.hire.Do(ctx, usecase.HireEmployeeInput{
		Code:    req.Msg.GetCode(),
		Name:    req.Msg.GetName(),
		Email:   req.Msg.GetEmail(),
		HiredOn: converters.ToTime(req.Msg.GetHiredOn()),
	})
	if err != nil {
		return nil, cerrors.Map(err)
	}

	return connect.NewResponse(&employeev1.HireEmployeeResponse{
		Employee: converter.EmployeeToEmployeev1(out.Employee),
	}), nil
}

func (h *Handler) GetEmployee(
	ctx context.Context,
	req *connect.Request[employeev1.GetEmployeeRequest],
) (*connect.Response[employeev1.GetEmployeeResponse], error) {
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.get.Do(ctx, usecase.GetEmployeeInput{ID: id})
	if err != nil {
		return nil, cerrors.Map(err)
	}

	return connect.NewResponse(&employeev1.GetEmployeeResponse{
		Employee: converter.EmployeeToEmployeev1(out.Employee),
	}), nil
}

func (h *Handler) ListEmployees(
	ctx context.Context,
	req *connect.Request[employeev1.ListEmployeesRequest],
) (*connect.Response[employeev1.ListEmployeesResponse], error) {
	pageSize := req.Msg.GetPageSize()
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	afterID := uuid.Nil
	if token := req.Msg.GetPageToken(); token != "" {
		id, err := uuid.Parse(token)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid page token"))
		}
		afterID = id
	}

	out, err := h.list.Do(ctx, usecase.ListEmployeesInput{
		AfterID:  afterID,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, cerrors.Map(err)
	}

	items := make([]*employeev1.Employee, len(out.Employees))
	for i := range out.Employees {
		items[i] = converter.EmployeeToEmployeev1(out.Employees[i])
	}

	nextPageToken := ""
	if len(out.Employees) == int(pageSize) {
		nextPageToken = out.Employees[len(out.Employees)-1].ID.String()
	}

	return connect.NewResponse(&employeev1.ListEmployeesResponse{
		Employees:     items,
		NextPageToken: nextPageToken,
	}), nil
}
