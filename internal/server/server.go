// Package server wires Connect service handlers into an HTTP server.
package server

import (
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"

	"github.com/mickamy/employee-management/gen/assignment/v1/assignmentv1connect"
	"github.com/mickamy/employee-management/gen/employee/v1/employeev1connect"
	"github.com/mickamy/employee-management/gen/organization/v1/organizationv1connect"
)

// Handler returns the HTTP handler serving every Connect service.
// Integration tests can mount it on an httptest.Server instead of running the binary.
func Handler() http.Handler {
	opts := connect.WithInterceptors(validate.NewInterceptor())

	mux := http.NewServeMux()
	mux.Handle(employeev1connect.NewEmployeeServiceHandler(
		employeev1connect.UnimplementedEmployeeServiceHandler{}, opts,
	))
	mux.Handle(assignmentv1connect.NewAssignmentCommandServiceHandler(
		assignmentv1connect.UnimplementedAssignmentCommandServiceHandler{}, opts,
	))
	mux.Handle(assignmentv1connect.NewAssignmentQueryServiceHandler(
		assignmentv1connect.UnimplementedAssignmentQueryServiceHandler{}, opts,
	))
	mux.Handle(organizationv1connect.NewOrganizationServiceHandler(
		organizationv1connect.UnimplementedOrganizationServiceHandler{}, opts,
	))
	return mux
}

// New returns an HTTP server serving Handler on addr.
//
// The server is meant to run behind a TLS-terminating reverse proxy
// (https-portal) and must not be exposed directly. Unencrypted HTTP/2 keeps
// gRPC clients, which require HTTP/2, working inside the network.
func New(addr string) *http.Server {
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	return &http.Server{
		Addr:              addr,
		Handler:           Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		Protocols:         protocols,
	}
}
