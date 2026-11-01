package auth

import (
	"context"
	"testing"

	"github.com/sawthorn/merch-wallet/internal/repository"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	getUser    func(context.Context, string) (*repository.User, error)
	createUser func(context.Context, string, string) (*repository.User, error)
}

func (r *mockUserRepo) GetUserByUsername(ctx context.Context, username string) (*repository.User, error) {
	return r.getUser(ctx, username)
}

func (r *mockUserRepo) CreateUser(ctx context.Context, username, password string) (*repository.User, error) {
	return r.createUser(ctx, username, password)
}

type mockTxManager struct{}

func (m *mockTxManager) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func TestLoginOrRegisterUser_LoginSuccess(t *testing.T) {
	username, password := "TestUser", "secretPass"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	user := repository.NewUser(
		repository.WithUsername(username),
		repository.WithPassword(string(hash)),
	)

	as := NewAuthService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				return user, nil
			},
			createUser: func(ctx context.Context, username, password string) (*repository.User, error) {
				t.Fatal("create should not be called in this testcase")
				return nil, nil
			},
		},
		&mockTxManager{},
	)

	_, err = as.LoginOrRegisterUser(context.Background(), username, password)
	require.NoError(t, err)
}

func TestLoginOrRegisterUser_RegisterSuccess(t *testing.T) {
	username, password := "TestUser", "secretPass"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	user := repository.NewUser(
		repository.WithUsername(username),
		repository.WithPassword(string(hash)),
		repository.WithBalance(uint(1000)),
	)

	as := NewAuthService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				return nil, repository.ErrUserNotFound
			},
			createUser: func(ctx context.Context, username string, password string) (*repository.User, error) {
				return user, nil
			},
		},
		&mockTxManager{},
	)

	_, err = as.LoginOrRegisterUser(context.Background(), username, password)
	require.NoError(t, err)
}

func TestLoginOrRegisterUser_PasswordMismatch(t *testing.T) {
	username, password := "TestUser", "secretPass"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	user := repository.NewUser(
		repository.WithUsername(username),
		repository.WithPassword(string(hash)),
	)

	as := NewAuthService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				return user, nil
			},
			createUser: func(ctx context.Context, username, password string) (*repository.User, error) {
				t.Fatal("create should not be called in this testcase")
				return nil, nil
			},
		},
		&mockTxManager{},
	)

	_, err = as.LoginOrRegisterUser(context.Background(), username, "wrongPass")
	require.ErrorIs(t, err, ErrPasswordMismatch)
}

func TestLoginOrRegisterUser_RetrySuccess(t *testing.T) {
	username, password := "TestUser", "secretPass"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	user := repository.NewUser(
		repository.WithUsername(username),
		repository.WithPassword(string(hash)),
	)

	isRetry := false
	as := NewAuthService(
		// simulates race to create one user between two requests
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				if isRetry == false {
					isRetry = true
					return nil, repository.ErrUserNotFound
				} else {
					return user, nil
				}
			},
			createUser: func(ctx context.Context, username, password string) (*repository.User, error) {
				return nil, repository.ErrUserAlreadyExists
			},
		},
		&mockTxManager{},
	)

	_, err = as.LoginOrRegisterUser(context.Background(), username, password)
	require.NoError(t, err)
}
