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

func TestTransferRepo_CreateTransfer(t *testing.T) {
	dsn := testutils.StartPostgres(t)
	testutils.RunMigrations(t, dsn)

	conn := database.NewPostgresConnectionWithDSN(dsn)
	ex := transactions.NewDBExecutor(conn)
	trRepo := repository.NewTransferRepository(ex)

	uid1 := 1
	uid2 := 2
	_, err := conn.ExecContext(
		context.Background(), `
		INSERT INTO users(id, username, password, balance) VALUES
		($1, $2, $3, 1000),
		($4, $5, $6, 1000)
		`,
		uid1, "u1", "u1pass",
		uid2, "u2", "u2pass",
	)
	require.NoError(t, err)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	testcases := []struct {
		name    string
		ctx     context.Context
		from    int
		to      int
		wantErr error
	}{
		{
			name: "Success",
			ctx:  context.Background(),
			from: uid1,
			to:   uid2,
		},
		{
			name:    "Fail_SenderNotFound",
			ctx:     context.Background(),
			from:    444,
			to:      uid2,
			wantErr: repository.ErrUserNotFound,
		},
		{
			name:    "Fail_ReceiverNotFound",
			ctx:     context.Background(),
			from:    uid1,
			to:      4444,
			wantErr: repository.ErrUserNotFound,
		},
		{
			name:    "Fail_Canceled",
			ctx:     canceledCtx,
			from:    uid1,
			to:      uid2,
			wantErr: context.Canceled,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err = trRepo.CreateTransfer(
				tc.ctx,
				tc.from,
				tc.to,
				10,
			)
			if tc.wantErr != nil {
				assert.ErrorIs(t, tc.wantErr, err)
			} else {
				assert.NoError(t, err)

				err = conn.QueryRowContext(
					context.Background(),
					`
					SELECT FROM transfers
					WHERE from_user_id = $1 AND to_user_id = $2
					`, tc.from, tc.to).Scan()
				assert.NoError(t, err)
			}
		})
	}
}

func TestTransferRepo_GetHistory_Both(t *testing.T) {
	dsn := testutils.StartPostgres(t)
	testutils.RunMigrations(t, dsn)

	conn := database.NewPostgresConnectionWithDSN(dsn)
	ex := transactions.NewDBExecutor(conn)
	trRepo := repository.NewTransferRepository(ex)

	uid1 := 1
	uid2 := 2
	uid3 := 3
	_, err := conn.ExecContext(
		context.Background(), `
		INSERT INTO users(id, username, password, balance) VALUES
		($1, $2, $3, 1000),
		($4, $5, $6, 1000),
		($7, $8, $9, 1000)
		`,
		uid1, "u1", "u1pass",
		uid2, "u2", "u2pass",
		uid3, "u3", "u3pass",
	)
	require.NoError(t, err)

	_, err = conn.ExecContext(
		context.Background(), `
		INSERT INTO transfers(from_user_id, to_user_id, amount) VALUES
		($1, $2, 1),
		($2, $3, 3),
		($3, $1, 5),
		($2, $1, 8),
		($1, $3, 13),
		($3, $2, 21)
		`, uid1, uid2, uid3)
	require.NoError(t, err)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	testcases := []struct {
		name    string
		ctx     context.Context
		userID  int
		histLen int
		wantErr error
	}{
		{
			name:    "Success",
			ctx:     context.Background(),
			userID:  uid1,
			histLen: 2,
		},
		{
			name:    "Fail_UserNotExists_EmptyHist",
			ctx:     context.Background(),
			userID:  444,
			histLen: 0,
		},
		{
			name:    "Fail_Canceled",
			ctx:     canceledCtx,
			userID:  uid1,
			wantErr: context.Canceled,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			sent, err := trRepo.GetSentHistory(
				tc.ctx,
				tc.userID,
			)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Len(t, sent, tc.histLen)
			}

			received, err := trRepo.GetReceivedHistory(
				tc.ctx,
				tc.userID,
			)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Len(t, received, tc.histLen)
			}
		})
	}
}
