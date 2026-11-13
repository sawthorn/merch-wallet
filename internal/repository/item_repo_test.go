package repository_test

import (
	"context"
	"testing"

	"github.com/sawthorn/merch-wallet/internal/database"
	"github.com/sawthorn/merch-wallet/internal/database/transactions"
	"github.com/sawthorn/merch-wallet/internal/repository"
	"github.com/sawthorn/merch-wallet/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItemRepo_GetItemByName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	dsn := testutils.StartPostgres(t)
	testutils.RunMigrations(t, dsn)

	conn := database.NewPostgresConnectionWithDSN(dsn)
	executor := transactions.NewDBExecutor(conn)
	itemRepo := repository.NewItemRepository(executor)

	itemName := ""
	itemCost := 100
	_, err := conn.ExecContext(
		context.Background(),
		`INSERT INTO items(name, cost)
		VALUES($1, $2)
		`, itemName, itemCost)
	require.NoError(t, err)

	item, err := itemRepo.GetItemByName(
		context.Background(),
		itemName,
	)

	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, itemName, item.Name())
	assert.Equal(t, itemCost, item.Cost())

	item, err = itemRepo.GetItemByName(
		context.Background(),
		"NotExists",
	)
	require.ErrorIs(t, repository.ErrItemNotFound, err)
	require.Nil(t, item)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	item, err = itemRepo.GetItemByName(ctx, itemName)
	require.ErrorIs(t, context.Canceled, err)
	require.Nil(t, item)
}
