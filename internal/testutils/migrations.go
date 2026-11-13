package testutils

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func RunMigrations(t *testing.T, dsn string) {
	ctx := context.Background()
	migrationsDir, err := filepath.Abs("../database/migrations")
	require.NoError(t, err)
	cmd := exec.CommandContext(
		ctx,
		"goose",
		"postgres",
		"-dir", migrationsDir,
		dsn,
		"up",
	)

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}
