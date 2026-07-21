package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/type/date"

	assignmentv1 "github.com/mickamy/employee-management/gen/assignment/v1"
	"github.com/mickamy/employee-management/gen/assignment/v1/assignmentv1connect"
	employeev1 "github.com/mickamy/employee-management/gen/employee/v1"
	"github.com/mickamy/employee-management/gen/employee/v1/employeev1connect"
	organizationv1 "github.com/mickamy/employee-management/gen/organization/v1"
	"github.com/mickamy/employee-management/gen/organization/v1/organizationv1connect"
	"github.com/mickamy/employee-management/internal/server"
)

const validUUID = "5b8f6f6e-7c1a-4b7e-9d2a-1f3e4d5c6b7a"

func TestHandler(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(server.Handler())
	t.Cleanup(ts.Close)

	tests := []struct {
		name string
		call func(ctx context.Context, hc *http.Client, url string) error
		want connect.Code
	}{
		{
			name: "employee: rejects invalid uuid",
			call: func(ctx context.Context, hc *http.Client, url string) error {
				c := employeev1connect.NewEmployeeServiceClient(hc, url)
				_, err := c.GetEmployee(ctx, connect.NewRequest(&employeev1.GetEmployeeRequest{Id: "not-a-uuid"}))
				return err
			},
			want: connect.CodeInvalidArgument,
		},
		{
			name: "employee: valid request reaches the stub",
			call: func(ctx context.Context, hc *http.Client, url string) error {
				c := employeev1connect.NewEmployeeServiceClient(hc, url)
				_, err := c.GetEmployee(ctx, connect.NewRequest(&employeev1.GetEmployeeRequest{Id: validUUID}))
				return err
			},
			want: connect.CodeUnimplemented,
		},
		{
			name: "employee: rejects zero-value hired_on",
			call: func(ctx context.Context, hc *http.Client, url string) error {
				c := employeev1connect.NewEmployeeServiceClient(hc, url)
				_, err := c.HireEmployee(ctx, connect.NewRequest(&employeev1.HireEmployeeRequest{
					Code:    "E0001",
					Name:    "Test Employee",
					Email:   "test@example.com",
					HiredOn: &date.Date{},
				}))
				return err
			},
			want: connect.CodeInvalidArgument,
		},
		{
			name: "assignment command: rejects empty request",
			call: func(ctx context.Context, hc *http.Client, url string) error {
				c := assignmentv1connect.NewAssignmentCommandServiceClient(hc, url)
				_, err := c.AssignEmployee(ctx, connect.NewRequest(&assignmentv1.AssignEmployeeRequest{}))
				return err
			},
			want: connect.CodeInvalidArgument,
		},
		{
			name: "assignment command: valid request reaches the stub",
			call: func(ctx context.Context, hc *http.Client, url string) error {
				c := assignmentv1connect.NewAssignmentCommandServiceClient(hc, url)
				_, err := c.AssignEmployee(ctx, connect.NewRequest(&assignmentv1.AssignEmployeeRequest{
					EmployeeId:   validUUID,
					DepartmentId: validUUID,
					Position:     assignmentv1.Position_POSITION_MEMBER,
					AssignedOn:   &date.Date{Year: 2026, Month: 4, Day: 1},
				}))
				return err
			},
			want: connect.CodeUnimplemented,
		},
		{
			name: "assignment query: rejects negative min_revision",
			call: func(ctx context.Context, hc *http.Client, url string) error {
				c := assignmentv1connect.NewAssignmentQueryServiceClient(hc, url)
				_, err := c.GetCurrentAssignment(ctx, connect.NewRequest(&assignmentv1.GetCurrentAssignmentRequest{
					EmployeeId:  validUUID,
					MinRevision: -1,
				}))
				return err
			},
			want: connect.CodeInvalidArgument,
		},
		{
			name: "assignment query: valid request reaches the stub",
			call: func(ctx context.Context, hc *http.Client, url string) error {
				c := assignmentv1connect.NewAssignmentQueryServiceClient(hc, url)
				_, err := c.GetCurrentAssignment(ctx, connect.NewRequest(&assignmentv1.GetCurrentAssignmentRequest{
					EmployeeId: validUUID,
				}))
				return err
			},
			want: connect.CodeUnimplemented,
		},
		{
			name: "organization: rejects empty name",
			call: func(ctx context.Context, hc *http.Client, url string) error {
				c := organizationv1connect.NewOrganizationServiceClient(hc, url)
				_, err := c.CreateDepartment(ctx, connect.NewRequest(&organizationv1.CreateDepartmentRequest{
					Code: "D0001",
					Name: "",
				}))
				return err
			},
			want: connect.CodeInvalidArgument,
		},
		{
			name: "organization: valid request reaches the stub",
			call: func(ctx context.Context, hc *http.Client, url string) error {
				c := organizationv1connect.NewOrganizationServiceClient(hc, url)
				_, err := c.CreateDepartment(ctx, connect.NewRequest(&organizationv1.CreateDepartmentRequest{
					Code: "D0001",
					Name: "Engineering",
				}))
				return err
			},
			want: connect.CodeUnimplemented,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.call(t.Context(), ts.Client(), ts.URL)
			if err == nil {
				t.Fatal("want error, got nil")
			}
			if got := connect.CodeOf(err); got != tt.want {
				t.Fatalf("connect.CodeOf(err) = %v, want %v (err: %v)", got, tt.want, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	srv := server.New(":0")

	if srv.Protocols == nil || !srv.Protocols.HTTP1() || !srv.Protocols.UnencryptedHTTP2() {
		t.Fatalf("Protocols = %v, want HTTP/1.1 and unencrypted HTTP/2 enabled", srv.Protocols)
	}
}
