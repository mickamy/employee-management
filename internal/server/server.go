// Package server wires Connect service handlers into an HTTP server.
package server

import (
	"net/http"
	"time"

	"connectrpc.com/connect"

	"github.com/mickamy/employee-management/gen/assignment/v1/assignmentv1connect"
	"github.com/mickamy/employee-management/gen/employee/v1/employeev1connect"
	"github.com/mickamy/employee-management/gen/organization/v1/organizationv1connect"
	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/server/interceptor"
)

// Handler returns the HTTP handler serving every Connect service. Services
// not yet implemented keep their generated Unimplemented handlers.
// Integration tests can mount it on an httptest.Server instead of running the binary.
func Handler(cfg di.Config, handlers Handlers) http.Handler {
	opts := []connect.HandlerOption{
		interceptor.Option(cfg),
	}

	mux := http.NewServeMux()

	mux.Handle(assignmentv1connect.NewAssignmentCommandServiceHandler(handlers.Assignment, opts...))
	mux.Handle(assignmentv1connect.NewAssignmentQueryServiceHandler(handlers.Assignment, opts...))
	mux.Handle(employeev1connect.NewEmployeeServiceHandler(handlers.Employee, opts...))
	mux.Handle(organizationv1connect.NewOrganizationServiceHandler(handlers.Organization, opts...))

	return mux
}

// New returns an HTTP server serving Handler on addr.
//
// The server is meant to run behind a TLS-terminating reverse proxy
// (https-portal) and must not be exposed directly. Unencrypted HTTP/2 keeps
// gRPC clients, which require HTTP/2, working inside the network.
func New(addr string, cfg di.Config, handlers Handlers) *http.Server {
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	return &http.Server{
		Addr:              addr,
		Handler:           Handler(cfg, handlers),
		ReadHeaderTimeout: 10 * time.Second,
		Protocols:         protocols,
	}
}
