// Package aerrors defines domain error kinds shared across feature.
// Repositories translate infrastructure errors into these; handlers map them
// onto transport codes.
package aerrors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("already exists")
)

// NotFound reports that the named entity does not exist.
func NotFound(entity string) error {
	return fmt.Errorf("%s: %w", entity, ErrNotFound)
}

// Conflict reports that the described value is already taken.
func Conflict(detail string) error {
	return fmt.Errorf("%s: %w", detail, ErrConflict)
}
