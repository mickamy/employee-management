package handler_test

import (
	"net/http"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/mickamy/contest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	organizationv1 "github.com/mickamy/employee-management/gen/organization/v1"
	"github.com/mickamy/employee-management/gen/organization/v1/organizationv1connect"
	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/feature/organization/fixture"
	"github.com/mickamy/employee-management/internal/feature/organization/handler"
	"github.com/mickamy/employee-management/internal/server/interceptor"
	"github.com/mickamy/employee-management/internal/test/tinfra"
)

func TestHandler_CreateAndRenameDepartment(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	department := fixture.Department()
	ct := contest.NewWith(t,
		contest.Bind(organizationv1connect.NewOrganizationServiceHandler)(handler.NewOrganization(infra)),
		connect.WithInterceptors(interceptor.NewInterceptors(*di.NewConfig())...),
	)
	var createOut organizationv1.CreateDepartmentResponse
	ct.
		Procedure(organizationv1connect.OrganizationServiceCreateDepartmentProcedure).
		In(&organizationv1.CreateDepartmentRequest{
			Code: department.Code,
			Name: department.Name,
		}).
		Do().
		ExpectStatus(http.StatusOK).
		Out(&createOut)
	require.NotEmpty(t, createOut.GetDepartment().GetId())
	require.Equal(t, department.Code, createOut.GetDepartment().GetCode())
	require.Equal(t, department.Name, createOut.GetDepartment().GetName())

	// act
	var renameOut organizationv1.RenameDepartmentResponse
	ct.
		Procedure(organizationv1connect.OrganizationServiceRenameDepartmentProcedure).
		In(&organizationv1.RenameDepartmentRequest{
			Id:   createOut.GetDepartment().GetId(),
			Name: "Renamed",
		}).
		Do().
		ExpectStatus(http.StatusOK).
		Out(&renameOut)

	// assert
	assert.Equal(t, createOut.GetDepartment().GetId(), renameOut.GetDepartment().GetId())
	assert.Equal(t, department.Code, renameOut.GetDepartment().GetCode())
	assert.Equal(t, "Renamed", renameOut.GetDepartment().GetName())
}

func TestHandler_RenameDepartment_NotFound(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)

	// act
	ct := contest.NewWith(t,
		contest.Bind(organizationv1connect.NewOrganizationServiceHandler)(handler.NewOrganization(infra)),
		connect.WithInterceptors(interceptor.NewInterceptors(*di.NewConfig())...),
	).
		Procedure(organizationv1connect.OrganizationServiceRenameDepartmentProcedure).
		In(&organizationv1.RenameDepartmentRequest{
			Id:   uuid.New().String(),
			Name: "Renamed",
		}).
		Do()

	// assert
	ct.ExpectStatus(http.StatusNotFound)
}

func TestHandler_ListDepartments(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	ct := contest.NewWith(t,
		contest.Bind(organizationv1connect.NewOrganizationServiceHandler)(handler.NewOrganization(infra)),
		connect.WithInterceptors(interceptor.NewInterceptors(*di.NewConfig())...),
	)
	// created in the opposite order of code, so the returned order can only
	// come from ORDER BY code.
	for _, code := range []string{"b-team", "a-team"} {
		ct.
			Procedure(organizationv1connect.OrganizationServiceCreateDepartmentProcedure).
			In(&organizationv1.CreateDepartmentRequest{
				Code: code,
				Name: "Team " + code,
			}).
			Do().
			ExpectStatus(http.StatusOK)
	}

	// act
	var out organizationv1.ListDepartmentsResponse
	ct.
		Procedure(organizationv1connect.OrganizationServiceListDepartmentsProcedure).
		In(&organizationv1.ListDepartmentsRequest{}).
		Do().
		ExpectStatus(http.StatusOK).
		Out(&out)

	// assert
	require.Len(t, out.GetDepartments(), 2)
	assert.Equal(t, "a-team", out.GetDepartments()[0].GetCode())
	assert.Equal(t, "b-team", out.GetDepartments()[1].GetCode())
}
