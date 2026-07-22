// Package tinfra provisions the application's Infra wired to an isolated,
// freshly migrated test database.
package tinfra

import (
	"testing"

	"github.com/mickamy/employee-management/internal/di"
	"github.com/mickamy/employee-management/internal/test/tdb"
)

// New returns an Infra backed by an isolated test database. Everything is
// cleaned up with the test.
func New(t *testing.T) di.Infra {
	t.Helper()

	d := tdb.New(t)
	return di.Infra{Writer: d.Writer, Reader: d.Reader}
}
