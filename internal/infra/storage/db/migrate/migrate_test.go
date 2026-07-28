package migrate_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/config"
	"github.com/mickamy/employee-management/internal/infra/storage/db"
	"github.com/mickamy/employee-management/internal/infra/storage/db/migrate"
)

func TestUp(t *testing.T) {
	t.Parallel()

	adminURL := config.ParseDatabase().AdminURL
	admin, err := db.New(t.Context(), adminURL)
	require.NoError(t, err)
	t.Cleanup(admin.Close)

	name := fmt.Sprintf("migrate_test_%d", time.Now().UnixNano())
	ident := pgx.Identifier{name}.Sanitize()
	_, err = admin.Exec(t.Context(), "CREATE DATABASE "+ident)
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx := context.WithoutCancel(t.Context())
		_, err := admin.Exec(ctx, "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)")
		assert.NoError(t, err)
	})

	pool, err := db.New(t.Context(), replaceDBName(t, adminURL, name))
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	require.NoError(t, migrate.Up(t.Context(), pool))

	var tables int
	row := pool.QueryRow(t.Context(),
		"SELECT count(*) FROM information_schema.tables WHERE table_name IN ('employees', 'employee_hires')")
	require.NoError(t, row.Scan(&tables))
	require.Equal(t, 2, tables)

	assert.NoError(t, migrate.Up(t.Context(), pool), "second run must be a no-op")
}

func replaceDBName(t *testing.T, raw, name string) string {
	t.Helper()

	u, err := url.Parse(raw)
	require.NoError(t, err)
	u.Path = "/" + name
	return u.String()
}
