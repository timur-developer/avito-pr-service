package postgres

import (
	"context"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

const (
	testDBName     = "testdb"
	testDBUser     = "testuser"
	testDBPassword = "testpass"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	_, currentFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(currentFile))))
	migrationsDir := filepath.Join(projectRoot, "migrations")

	t.Logf("Looking for migrations in: %s", migrationsDir)

	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		t.Fatalf("migrations directory not found: %s", migrationsDir)
	}

	migrationsSource := "file://" + filepath.ToSlash(migrationsDir)
	t.Logf("Migrations source URL: %s", migrationsSource)

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:15-alpine",
		postgres.WithDatabase(testDBName),
		postgres.WithUsername(testDBUser),
		postgres.WithPassword(testDBPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	m, err := migrate.New(migrationsSource, dsn)
	require.NoError(t, err)

	err = m.Up()
	require.NoError(t, err)

	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	dbPool, err := pgxpool.NewWithConfig(ctx, config)
	require.NoError(t, err)

	cleanup := func() {
		dbPool.Close()
		_ = pgContainer.Terminate(ctx)
	}

	return dbPool, cleanup
}
