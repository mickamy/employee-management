package handler

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	organizationv1 "github.com/mickamy/employee-management/gen/organization/v1"
	"github.com/mickamy/employee-management/gen/organization/v1/organizationv1connect"
	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/errors/cerrors"
	"github.com/mickamy/employee-management/internal/feature/organization/converter"
	"github.com/mickamy/employee-management/internal/feature/organization/usecase"
)

type Organization struct {
	_      di.Infra                  `inject:"embed"`
	create *usecase.CreateDepartment `inject:""`
	rename *usecase.RenameDepartment `inject:""`
	list   *usecase.ListDepartments  `inject:""`
}

var _ organizationv1connect.OrganizationServiceHandler = (*Organization)(nil)

func (h *Organization) CreateDepartment(
	ctx context.Context,
	req *connect.Request[organizationv1.CreateDepartmentRequest],
) (*connect.Response[organizationv1.CreateDepartmentResponse], error) {
	out, err := h.create.Do(ctx, usecase.CreateDepartmentInput{
		Code: req.Msg.GetCode(),
		Name: req.Msg.GetName(),
	})
	if err != nil {
		return nil, cerrors.Map(err)
	}

	return connect.NewResponse(&organizationv1.CreateDepartmentResponse{
		Department: converter.DepartmentToOrganizationv1(out.Department),
	}), nil
}

func (h *Organization) RenameDepartment(
	ctx context.Context,
	req *connect.Request[organizationv1.RenameDepartmentRequest],
) (*connect.Response[organizationv1.RenameDepartmentResponse], error) {
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.rename.Do(ctx, usecase.RenameDepartmentInput{
		ID:   id,
		Name: req.Msg.GetName(),
	})
	if err != nil {
		return nil, cerrors.Map(err)
	}

	return connect.NewResponse(&organizationv1.RenameDepartmentResponse{
		Department: converter.DepartmentToOrganizationv1(out.Department),
	}), nil
}

func (h *Organization) ListDepartments(
	ctx context.Context,
	req *connect.Request[organizationv1.ListDepartmentsRequest],
) (*connect.Response[organizationv1.ListDepartmentsResponse], error) {
	out, err := h.list.Do(ctx, usecase.ListDepartmentsInput{})
	if err != nil {
		return nil, cerrors.Map(err)
	}

	items := make([]*organizationv1.Department, len(out.Departments))
	for i := range out.Departments {
		items[i] = converter.DepartmentToOrganizationv1(out.Departments[i])
	}

	return connect.NewResponse(&organizationv1.ListDepartmentsResponse{
		Departments: items,
	}), nil
}
