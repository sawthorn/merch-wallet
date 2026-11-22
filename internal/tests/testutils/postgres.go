package testutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func StartPostgresForTest(t *testing.T) (dsn string, restoreFn func() error) {
	t.Helper()
	ctx := context.Background()
	ctr, err := postgres.Run(
		ctx,
		"postgres:18",
		postgres.WithDatabase("mwall-test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	dsn, err = ctr.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	err = RunMigrations(dsn)
	require.NoError(t, err)

	return dsn, nil
}
