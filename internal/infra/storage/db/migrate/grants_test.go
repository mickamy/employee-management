package migrate_test

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/test/tdb"
)

// eventTables lists every INSERT-only event table. New event tables must be
// added here so their immutability stays pinned by tests.
var eventTables = []struct {
	table  string
	column string
}{
	{table: "employee_hires", column: "hired_on"},
	{table: "assignment_events", column: "event_type"},
}

const insufficientPrivilege = "42501"

func TestEventTablesAreInsertOnly(t *testing.T) {
	t.Parallel()

	db := tdb.New(t)
	for _, target := range eventTables {
		for _, stmt := range []string{
			fmt.Sprintf("UPDATE %s SET %s = %s WHERE false", target.table, target.column, target.column),
			fmt.Sprintf("DELETE FROM %s WHERE false", target.table),
		} {
			_, err := db.Writer.Exec(t.Context(), stmt)
			require.Error(t, err, stmt)
			var pgErr *pgconn.PgError
			require.ErrorAs(t, err, &pgErr, stmt)
			assert.Equal(t, insufficientPrivilege, pgErr.Code, stmt)
		}
	}
}

func TestReaderCannotWrite(t *testing.T) {
	t.Parallel()

	db := tdb.New(t)
	for _, target := range eventTables {
		stmt := fmt.Sprintf("DELETE FROM %s WHERE false", target.table)
		_, err := db.Reader.Exec(t.Context(), stmt)
		require.Error(t, err, stmt)
		var pgErr *pgconn.PgError
		require.ErrorAs(t, err, &pgErr, stmt)
		assert.Equal(t, insufficientPrivilege, pgErr.Code, stmt)
	}
}
