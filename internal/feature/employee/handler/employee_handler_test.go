package handler_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/mickamy/contest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/type/date"

	"github.com/mickamy/employee-management/gen/employee/v1/employeev1connect"
	"github.com/mickamy/employee-management/internal/feature/employee/fixture"
	"github.com/mickamy/employee-management/internal/feature/employee/model"
	"github.com/mickamy/employee-management/internal/lib/converters"
	"github.com/mickamy/employee-management/internal/server/interceptor"

	employeev1 "github.com/mickamy/employee-management/gen/employee/v1"
	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/feature/employee/handler"
	"github.com/mickamy/employee-management/internal/test/tinfra"
)

func TestHandler_HireAndGet(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	hired := fixture.Employee(func(m *model.Employee) {
		m.HiredOn = m.HiredOn.Truncate(24 * time.Hour)
	})
	ct := contest.NewWith(t,
		contest.Bind(employeev1connect.NewEmployeeServiceHandler)(handler.NewEmployee(infra)),
		connect.WithInterceptors(interceptor.NewInterceptors(*di.NewConfig())...),
	)
	var hireOut employeev1.HireEmployeeResponse
	ct.
		Procedure(employeev1connect.EmployeeServiceHireEmployeeProcedure).
		In(&employeev1.HireEmployeeRequest{
			Code:    hired.Code,
			Name:    hired.Name,
			Email:   hired.Email,
			HiredOn: converters.ToDate(hired.HiredOn),
		}).
		Do().
		ExpectStatus(http.StatusOK).
		Out(&hireOut)
	require.Equal(t, hired.Code, hireOut.GetEmployee().GetCode())
	require.Equal(t, hired.Name, hireOut.GetEmployee().GetName())
	require.Equal(t, hired.Email, hireOut.GetEmployee().GetEmail())
	require.Equal(t, hired.HiredOn, converters.ToTime(hireOut.GetEmployee().GetHiredOn()))

	// act
	var getOut employeev1.GetEmployeeResponse
	ct.
		Procedure(employeev1connect.EmployeeServiceGetEmployeeProcedure).
		In(&employeev1.GetEmployeeRequest{
			Id: hireOut.GetEmployee().GetId(),
		}).
		Do().
		ExpectStatus(http.StatusOK).
		Out(&getOut)

	// assert
	assert.Equal(t, hired.Code, getOut.GetEmployee().GetCode())
	assert.Equal(t, hired.Name, getOut.GetEmployee().GetName())
	assert.Equal(t, hired.Email, getOut.GetEmployee().GetEmail())
	assert.Equal(t, hired.HiredOn, converters.ToTime(getOut.GetEmployee().GetHiredOn()))
}

func TestHandler_GetNotFound(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)

	// act
	ct := contest.NewWith(t,
		contest.Bind(employeev1connect.NewEmployeeServiceHandler)(handler.NewEmployee(infra)),
		connect.WithInterceptors(interceptor.NewInterceptors(*di.NewConfig())...),
	).
		Procedure(employeev1connect.EmployeeServiceGetEmployeeProcedure).
		In(&employeev1.GetEmployeeRequest{
			Id: uuid.New().String(),
		}).
		Do()

	// assert
	ct.ExpectStatus(http.StatusNotFound)
}

func TestHandler_ListPagination(t *testing.T) {
	t.Parallel()

	// arrange
	infra := tinfra.New(t)
	ct := contest.NewWith(t,
		contest.Bind(employeev1connect.NewEmployeeServiceHandler)(handler.NewEmployee(infra)),
		connect.WithInterceptors(interceptor.NewInterceptors(*di.NewConfig())...),
	)
	for i := range 3 {
		ct.
			Procedure(employeev1connect.EmployeeServiceHireEmployeeProcedure).
			In(&employeev1.HireEmployeeRequest{
				Code:    fmt.Sprintf("E%04d", i),
				Name:    fmt.Sprintf("Employee %d", i),
				Email:   fmt.Sprintf("employee%d@example.com", i),
				HiredOn: &date.Date{Year: 2026, Month: 4, Day: 1},
			}).
			Do().
			ExpectStatus(http.StatusOK)
	}

	// act
	var firstOut employeev1.ListEmployeesResponse
	ct.
		Procedure(employeev1connect.EmployeeServiceListEmployeesProcedure).
		In(&employeev1.ListEmployeesRequest{
			PageSize: 2,
		}).
		Do().
		ExpectStatus(http.StatusOK).
		Out(&firstOut)
	require.Len(t, firstOut.GetEmployees(), 2)
	require.NotEmpty(t, firstOut.GetNextPageToken())
	var secondOut employeev1.ListEmployeesResponse
	ct.
		Procedure(employeev1connect.EmployeeServiceListEmployeesProcedure).
		In(&employeev1.ListEmployeesRequest{
			PageSize:  2,
			PageToken: firstOut.GetNextPageToken(),
		}).
		Do().
		ExpectStatus(http.StatusOK).
		Out(&secondOut)

	// assert
	assert.Len(t, secondOut.GetEmployees(), 1)
	assert.Empty(t, secondOut.GetNextPageToken())
}
