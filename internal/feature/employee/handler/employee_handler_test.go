package handler_test

import (
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/type/date"

	employeev1 "github.com/mickamy/employee-management/gen/employee/v1"
	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/feature/employee/handler"
	"github.com/mickamy/employee-management/internal/test/tinfra"
)

func newHandler(infra di.Infra) *handler.Handler {
	return handler.NewHandler(infra)
}

func TestHandlerHireAndGet(t *testing.T) {
	t.Parallel()

	infra := tinfra.New(t)
	h := newHandler(infra)

	hiredOn := &date.Date{Year: 2026, Month: 4, Day: 1}
	res, err := h.HireEmployee(t.Context(), connect.NewRequest(&employeev1.HireEmployeeRequest{
		Code:    "E0001",
		Name:    "Alice Tanaka",
		Email:   "alice@example.com",
		HiredOn: hiredOn,
	}))
	require.NoError(t, err)

	created := res.Msg.GetEmployee()
	_, err = uuid.Parse(created.GetId())
	require.NoError(t, err)

	got, err := h.GetEmployee(t.Context(), connect.NewRequest(&employeev1.GetEmployeeRequest{
		Id: created.GetId(),
	}))
	require.NoError(t, err)

	fetched := got.Msg.GetEmployee()
	require.Equal(t, "E0001", fetched.GetCode())
	require.Equal(t, "Alice Tanaka", fetched.GetName())
	require.Equal(t, "alice@example.com", fetched.GetEmail())
	require.Equal(t, hiredOn.GetYear(), fetched.GetHiredOn().GetYear())
	require.Equal(t, hiredOn.GetMonth(), fetched.GetHiredOn().GetMonth())
	require.Equal(t, hiredOn.GetDay(), fetched.GetHiredOn().GetDay())
}

func TestHandlerHireDuplicateCode(t *testing.T) {
	t.Parallel()

	infra := tinfra.New(t)
	h := newHandler(infra)

	_, err := h.HireEmployee(t.Context(), connect.NewRequest(hireRequest(1)))
	require.NoError(t, err)

	dup := hireRequest(2)
	dup.Code = hireRequest(1).GetCode()
	_, err = h.HireEmployee(t.Context(), connect.NewRequest(dup))
	require.Error(t, err)
	require.Equal(t, connect.CodeAlreadyExists, connect.CodeOf(err))
}

func TestHandlerGetNotFound(t *testing.T) {
	t.Parallel()

	infra := tinfra.New(t)
	h := newHandler(infra)

	_, err := h.GetEmployee(t.Context(), connect.NewRequest(&employeev1.GetEmployeeRequest{
		Id: uuid.NewString(),
	}))
	require.Error(t, err)
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

func TestHandlerListPagination(t *testing.T) {
	t.Parallel()

	infra := tinfra.New(t)
	h := newHandler(infra)

	for i := 1; i <= 3; i++ {
		_, err := h.HireEmployee(t.Context(), connect.NewRequest(hireRequest(i)))
		require.NoError(t, err)
	}

	first, err := h.ListEmployees(t.Context(), connect.NewRequest(&employeev1.ListEmployeesRequest{
		PageSize: 2,
	}))
	require.NoError(t, err)
	require.Len(t, first.Msg.GetEmployees(), 2)
	require.Equal(t, "E0001", first.Msg.GetEmployees()[0].GetCode())
	require.Equal(t, "E0002", first.Msg.GetEmployees()[1].GetCode())
	require.NotEmpty(t, first.Msg.GetNextPageToken())

	second, err := h.ListEmployees(t.Context(), connect.NewRequest(&employeev1.ListEmployeesRequest{
		PageSize:  2,
		PageToken: first.Msg.GetNextPageToken(),
	}))
	require.NoError(t, err)
	require.Len(t, second.Msg.GetEmployees(), 1)
	require.Equal(t, "E0003", second.Msg.GetEmployees()[0].GetCode())
	require.Empty(t, second.Msg.GetNextPageToken())
}

func TestHireEventsAreInsertOnly(t *testing.T) {
	t.Parallel()

	infra := tinfra.New(t)
	h := newHandler(infra)

	_, err := h.HireEmployee(t.Context(), connect.NewRequest(hireRequest(1)))
	require.NoError(t, err)

	_, err = infra.Writer.Exec(t.Context(), "UPDATE employee_hires SET hired_on = hired_on + INTERVAL '1 day'")
	require.ErrorContains(t, err, "permission denied")

	_, err = infra.Writer.Exec(t.Context(), "DELETE FROM employee_hires")
	require.ErrorContains(t, err, "permission denied")
}

func hireRequest(n int) *employeev1.HireEmployeeRequest {
	return &employeev1.HireEmployeeRequest{
		Code:    fmt.Sprintf("E%04d", n),
		Name:    fmt.Sprintf("Employee %d", n),
		Email:   fmt.Sprintf("employee%d@example.com", n),
		HiredOn: &date.Date{Year: 2026, Month: 4, Day: 1},
	}
}
