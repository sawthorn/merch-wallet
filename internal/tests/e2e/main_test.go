package e2e

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/sawthorn/merch-wallet/internal/app"
	"github.com/sawthorn/merch-wallet/internal/database"
	"github.com/sawthorn/merch-wallet/internal/tests/testutils"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type TestEnv struct {
	serverURL string
	resetDB   func() error
}

var env *TestEnv

func TestMain(m *testing.M) {
	ctr, dsn, err := StartPostgres()
	if err != nil {
		log.Fatalf("E2E postgres start failed: %v", err)
	}
	resetFn, err := MakeCleanSnapshot(ctr)
	if err != nil {
		log.Fatalf("E2E make clean snapshot failed: %v", err)
	}

	conn := database.NewPostgresConnectionWithDSN(dsn)
	app := app.NewApp(conn)
	server := httptest.NewServer(app.SetupRouter())

	env = &TestEnv{
		serverURL: server.URL,
		resetDB:   resetFn,
	}

	code := m.Run()
	server.Close()
	err = testcontainers.TerminateContainer(ctr)
	if err != nil {
		log.Fatalf("E2E terminate container failed: %v", err)
	}

	os.Exit(code)
}

func StartPostgres() (ctr *postgres.PostgresContainer, dsn string, err error) {
	ctx := context.Background()
	ctr, err = postgres.Run(
		ctx,
		"postgres:18",
		postgres.WithDatabase("mwall-test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, "", err
	}

	dsn, err = ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, "", err
	}

	err = testutils.RunMigrations(dsn)

	return ctr, dsn, err
}

func MakeCleanSnapshot(ctr *postgres.PostgresContainer) (resetFn func() error, err error) {
	ctx := context.Background()
	err = ctr.Snapshot(
		ctx,
		postgres.WithSnapshotName("clean"),
	)
	if err != nil {
		return nil, err
	}

	resetFn = func() error {
		return ctr.Restore(ctx, postgres.WithSnapshotName("clean"))
	}

	return resetFn, nil
}
