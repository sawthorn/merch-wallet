package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func StartPostgres(t *testing.T) (dsn string) {
	t.Helper()
	ctx := context.Background()
	container, err := postgres.Run(
		ctx,
		"postgres:18",
		postgres.WithDatabase("mwall-test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	dsn, err = container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	return dsn
}
