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

func TestPurchaseRepo_CreatePurchase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	dsn := testutils.StartPostgres(t)
	testutils.RunMigrations(t, dsn)

	conn := database.NewPostgresConnectionWithDSN(dsn)
	ex := transactions.NewDBExecutor(conn)
	purchaseRepo := repository.NewPurchaseRepository(ex)

	userID := 1
	userName, userPass := "testuser", "testpass"
	_, err := conn.ExecContext(
		context.Background(),
		`
		INSERT INTO users(id, username, password, balance)
		VALUES($1, $2, $3, 1000)
		`, userID, userName, userPass)
	require.NoError(t, err)

	testItemID := 1
	testItemName, testItemCost := "testitem", 10
	_, err = conn.ExecContext(
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
			userID:  444,
			itemID:  testItemID,
			wantErr: repository.ErrUserNotFound,
		},
		{
			name:    "Fail_ItemNotFound",
			ctx:     context.Background(),
			userID:  userID,
			itemID:  444,
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
			err = purchaseRepo.CreatePurchase(
				tc.ctx,
				tc.userID,
				tc.itemID,
			)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)

				// persistence check
				err = conn.QueryRowContext(
					context.Background(),
					`
					SELECT FROM purchases
					WHERE user_id = $1 AND item_id = $2
					`, tc.userID, tc.itemID).Scan()
				assert.NoError(t, err)
			}
		})
	}
}
