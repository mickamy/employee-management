// Package cerrors maps domain errors onto Connect error codes.
package cerrors

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/mickamy/employee-management/internal/errors/aerrors"
)

// Map translates a domain error into a Connect error. Unrecognized errors
// become internal errors.
func Map(err error) *connect.Error {
	switch {
	case errors.Is(err, aerrors.ErrConflict):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, aerrors.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
