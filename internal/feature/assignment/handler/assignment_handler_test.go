package handler_test

import (
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/mickamy/contest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/type/date"

	assignmentv1 "github.com/mickamy/employee-management/gen/assignment/v1"
	"github.com/mickamy/employee-management/gen/assignment/v1/assignmentv1connect"
	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/feature/assignment/handler"
	efixture "github.com/mickamy/employee-management/internal/feature/employee/fixture"
	eusecase "github.com/mickamy/employee-management/internal/feature/employee/usecase"
	ofixture "github.com/mickamy/employee-management/internal/feature/organization/fixture"
	ousecase "github.com/mickamy/employee-management/internal/feature/organization/usecase"
	"github.com/mickamy/employee-management/internal/server/interceptor"
	"github.com/mickamy/employee-management/internal/test/tinfra"
)

type services struct {
	command *contest.Client
	query   *contest.Client
}

func newServices(t *testing.T, infra di.Infra) services {
	t.Helper()

	h := handler.NewAssignment(infra)
	opts := connect.WithInterceptors(interceptor.NewInterceptors(*di.NewConfig())...)
	return services{
		command: contest.NewWith(t, contest.Bind(assignmentv1connect.NewAssignmentCommandServiceHandler)(h), opts),
		query:   contest.NewWith(t, contest.Bind(assignmentv1connect.NewAssignmentQueryServiceHandler)(h), opts),
	}
}

func arrangeEmployeeAndDepartment(t *testing.T, infra di.Infra) (uuid.UUID, uuid.UUID) {
	t.Helper()

	e := efixture.Employee()
	hired, err := eusecase.NewHireEmployee(infra).Do(t.Context(), eusecase.HireEmployeeInput{
		Code:    e.Code,
		Name:    e.Name,
		Email:   e.Email,
		HiredOn: e.HiredOn.Truncate(24 * time.Hour),
	})
	require.NoError(t, err)
	d := ofixture.Department()
	created, err := ousecase.NewCreateDepartment(infra).Do(t.Context(), ousecase.CreateDepartmentInput{
		Code: d.Code,
		Name: d.Name,
	})
	require.NoError(t, err)
	return hired.Employee.ID, created.Department.ID
}

func TestHandler_AssignReleaseFlow(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	employeeID, departmentID := arrangeEmployeeAndDepartment(t, infra)
	svc := newServices(t, infra)

	// act: assign, then read the write back immediately
	var assignOut assignmentv1.AssignEmployeeResponse
	svc.command.
		Procedure(assignmentv1connect.AssignmentCommandServiceAssignEmployeeProcedure).
		In(&assignmentv1.AssignEmployeeRequest{
			EmployeeId:   employeeID.String(),
			DepartmentId: departmentID.String(),
			Position:     assignmentv1.Position_POSITION_MEMBER,
			AssignedOn:   &date.Date{Year: 2026, Month: 4, Day: 1},
		}).
		Do().
		ExpectStatus(http.StatusOK).
		Out(&assignOut)
	require.NotEmpty(t, assignOut.GetAssignment().GetId())

	var currentOut assignmentv1.GetCurrentAssignmentResponse
	svc.query.
		Procedure(assignmentv1connect.AssignmentQueryServiceGetCurrentAssignmentProcedure).
		In(&assignmentv1.GetCurrentAssignmentRequest{
			EmployeeId: employeeID.String(),
		}).
		Do().
		ExpectStatus(http.StatusOK).
		Out(&currentOut)
	require.Equal(t, assignOut.GetAssignment().GetId(), currentOut.GetAssignment().GetId())

	var releaseOut assignmentv1.ReleaseAssignmentResponse
	svc.command.
		Procedure(assignmentv1connect.AssignmentCommandServiceReleaseAssignmentProcedure).
		In(&assignmentv1.ReleaseAssignmentRequest{
			AssignmentId: assignOut.GetAssignment().GetId(),
			ReleasedOn:   &date.Date{Year: 2026, Month: 9, Day: 30},
		}).
		Do().
		ExpectStatus(http.StatusOK).
		Out(&releaseOut)

	// assert: history reflects the release
	var historyOut assignmentv1.ListAssignmentHistoryResponse
	svc.query.
		Procedure(assignmentv1connect.AssignmentQueryServiceListAssignmentHistoryProcedure).
		In(&assignmentv1.ListAssignmentHistoryRequest{
			EmployeeId: employeeID.String(),
		}).
		Do().
		ExpectStatus(http.StatusOK).
		Out(&historyOut)
	require.Len(t, historyOut.GetAssignments(), 1)
	assert.NotNil(t, historyOut.GetAssignments()[0].GetReleasedOn())
}

func TestHandler_RevokeAssignment_AlreadyInEffect(t *testing.T) {
	t.Parallel()

	// arrange: an assignment already in effect cannot be revoked
	infra := tinfra.New(t)
	employeeID, departmentID := arrangeEmployeeAndDepartment(t, infra)
	svc := newServices(t, infra)
	var assignOut assignmentv1.AssignEmployeeResponse
	svc.command.
		Procedure(assignmentv1connect.AssignmentCommandServiceAssignEmployeeProcedure).
		In(&assignmentv1.AssignEmployeeRequest{
			EmployeeId:   employeeID.String(),
			DepartmentId: departmentID.String(),
			Position:     assignmentv1.Position_POSITION_MEMBER,
			AssignedOn:   &date.Date{Year: 2026, Month: 4, Day: 1},
		}).
		Do().
		ExpectStatus(http.StatusOK).
		Out(&assignOut)

	// act
	ct := svc.command.
		Procedure(assignmentv1connect.AssignmentCommandServiceRevokeAssignmentProcedure).
		In(&assignmentv1.RevokeAssignmentRequest{
			AssignmentId: assignOut.GetAssignment().GetId(),
			Reason:       "too late",
		}).
		Do()

	// assert: connect maps CodeFailedPrecondition to HTTP 400
	ct.ExpectStatus(http.StatusBadRequest)
}

func TestHandler_GetCurrentAssignment_NotFound(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	svc := newServices(t, infra)

	// act
	ct := svc.query.
		Procedure(assignmentv1connect.AssignmentQueryServiceGetCurrentAssignmentProcedure).
		In(&assignmentv1.GetCurrentAssignmentRequest{
			EmployeeId: uuid.New().String(),
		}).
		Do()

	// assert
	ct.ExpectStatus(http.StatusNotFound)
}
