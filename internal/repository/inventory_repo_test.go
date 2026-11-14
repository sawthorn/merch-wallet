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

func TestInventoryRepo_AddItemToUserInventory(t *testing.T) {
	dsn := testutils.StartPostgres(t)
	testutils.RunMigrations(t, dsn)

	db := database.NewPostgresConnectionWithDSN(dsn)
	ex := transactions.NewDBExecutor(db)
	invRepo := repository.NewInventoryRepository(ex)

	userID := 1
	userName, userPass := "testuser", "testpass"
	_, err := db.ExecContext(
		context.Background(),
		`
		INSERT INTO users(id, username, password, balance)
		VALUES($1, $2, $3, 1000)
		`, userID, userName, userPass)
	require.NoError(t, err)

	testItemID := 1
	testItemName, testItemCost := "testitem", 10
	_, err = db.ExecContext(
		context.Background(),
		`INSERT INTO items(id, name, cost)
		VALUES($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
		name = excluded.name,
		cost = excluded.cost`, testItemID, testItemName, testItemCost)
	require.NoError(t, err)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	testcases := []struct {
		name    string
		ctx     context.Context
		userID  int
		itemID  int
		wantErr error
	}{
		{
			name:   "Success",
			ctx:    context.Background(),
			userID: userID,
			itemID: testItemID,
		},
		{
			name:    "Fail_UserNotFound",
			ctx:     context.Background(),
			userID:  3424908,
			itemID:  testItemID,
			wantErr: repository.ErrUserNotFound,
		},
		{
			name:    "Fail_ItemNotFound",
			ctx:     context.Background(),
			userID:  userID,
			itemID:  309183,
			wantErr: repository.ErrItemNotFound,
		},
		{
			name:    "Fail_Canceled",
			ctx:     canceledCtx,
			userID:  userID,
			itemID:  testItemID,
			wantErr: context.Canceled,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err = invRepo.AddItemToUserInventory(
				tc.ctx,
				tc.userID, tc.itemID,
			)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInventoryRepo_GetUserInventory(t *testing.T) {
	dsn := testutils.StartPostgres(t)
	testutils.RunMigrations(t, dsn)

	db := database.NewPostgresConnectionWithDSN(dsn)
	ex := transactions.NewDBExecutor(db)
	invRepo := repository.NewInventoryRepository(ex)

	userID := 1
	userName, userPass := "testuser", "testpass"
	_, err := db.ExecContext(
		context.Background(),
		`
		INSERT INTO users(id, username, password, balance)
		VALUES($1, $2, $3, 1000)
		`, userID, userName, userPass)
	require.NoError(t, err)

	_, err = db.ExecContext(
		context.Background(),
		`INSERT INTO items(id, name, cost) VALUES
		(1, 'test1', 10),
		(2, 'test2', 20)
		ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		cost = excluded.cost
		`)
	require.NoError(t, err)

	_, err = db.ExecContext(
		context.Background(),
		`
		INSERT INTO inventory(user_id, item_id, quantity)
		VALUES($1, 1, 1),
		($1, 2, 1)
		`, userID)
	require.NoError(t, err)

	successInv := []repository.InventoryItem{
		repository.NewInventoryItem("test1", 1),
		repository.NewInventoryItem("test2", 1),
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	testcases := []struct {
		name    string
		ctx     context.Context
		userID  int
		wantInv []repository.InventoryItem
		wantErr error
	}{
		{
			name:    "Success",
			ctx:     context.Background(),
			userID:  userID,
			wantInv: successInv,
		},
		{
			name:   "Fail_UserNotExists_EmptyInv",
			ctx:    context.Background(),
			userID: 3424908,
		},
		{
			name:    "Fail_Canceled",
			ctx:     canceledCtx,
			userID:  userID,
			wantErr: context.Canceled,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			inv, err := invRepo.GetUserInventory(
				tc.ctx,
				tc.userID,
			)
			if tc.wantErr != nil {
				assert.ErrorIs(t, tc.wantErr, err)
				assert.Nil(t, inv)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantInv, inv)
			}
		})
	}
}
