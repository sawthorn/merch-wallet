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

func TestUserRepo_GetByUsername(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
	dsn := testutils.StartPostgres(t)
	testutils.RunMigrations(t, dsn)

	db := database.NewPostgresConnectionWithDSN(dsn)
	executor := transactions.NewDBExecutor(db)
	userRepo := repository.NewUserRepository(executor)

	testUsername := "testuser"
	testPassword := "testPass"
	_, err := db.Exec(`
		INSERT INTO users(username, password, balance)
		VALUES($1, $2, 1000)
		`, testUsername, testPassword)
	require.NoError(t, err)

	user, err := userRepo.GetUserByUsername(context.Background(), testUsername)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, testUsername, user.Username())

	user, err = userRepo.GetUserByUsername(
		context.Background(),
		"NotExists",
	)

	require.ErrorIs(t, repository.ErrUserNotFound, err)
	require.Nil(t, user)
}

func TestUserRepo_CreateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
	dsn := testutils.StartPostgres(t)
	testutils.RunMigrations(t, dsn)

	db := database.NewPostgresConnectionWithDSN(dsn)
	executor := transactions.NewDBExecutor(db)
	userRepo := repository.NewUserRepository(executor)

	testUsername := "testuser"
	testPassword := "testpass"
	_, err := db.Exec(`
		INSERT INTO users(username, password, balance)
		VALUES($1, $2, 1000)
		`, testUsername, testPassword)
	require.NoError(t, err)

	testNewUsername := "testnewuser"
	testNewPassword := "testnewpass"

	user, err := userRepo.CreateUser(
		context.Background(),
		testNewUsername,
		testNewPassword,
	)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, testNewUsername, user.Username())

	user, err = userRepo.CreateUser(
		context.Background(),
		"NotExists",
		"",
	)

	require.ErrorIs(t, err, repository.ErrEmptyPassword)
	assert.Nil(t, user)

	user, err = userRepo.CreateUser(
		context.Background(),
		testUsername,
		testPassword,
	)

	require.ErrorIs(t, err, repository.ErrUserAlreadyExists)
	assert.Nil(t, user)
}

func TestUserRepo_ChangeBalance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
	dsn := testutils.StartPostgres(t)
	testutils.RunMigrations(t, dsn)

	db := database.NewPostgresConnectionWithDSN(dsn)
	executor := transactions.NewDBExecutor(db)
	userRepo := repository.NewUserRepository(executor)

	testUserID := 1
	testUsername := "testuser"
	testPassword := "testpass"
	_, err := db.Exec(`
		INSERT INTO users(id, username, password, balance)
		VALUES($1, $2, $3, 1000)
		`, testUserID, testUsername, testPassword)
	require.NoError(t, err)

	testcases := []struct {
		name    string
		userID  int
		delta   int
		wantErr error
	}{
		{
			name:   "Success_DeltaPositive",
			userID: testUserID,
			delta:  10,
		},
		{
			name:   "Success_DeltaNegative",
			userID: testUserID,
			delta:  -10,
		},
		{
			name:    "Fail_InsufficientBalance",
			userID:  testUserID,
			delta:   -10000000,
			wantErr: repository.ErrInsufficientBalance,
		},
		{
			name:    "Fail_IncorrectID",
			userID:  21134,
			delta:   -10000000,
			wantErr: repository.ErrUserNotFound,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := userRepo.ChangeBalance(
				context.Background(),
				tc.userID,
				tc.delta,
			)

			if tc.wantErr != nil {
				assert.ErrorIs(t, tc.wantErr, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
