package testutils

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func RunMigrations(dsn string) error {
	migrationsDir, err := filepath.Abs("../../database/migrations")
	if err != nil {
		return err
	}
	cmd := exec.Command(
		"goose",
		"postgres",
		"-dir", migrationsDir,
		dsn,
		"up",
	)

	_, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
