package transfer

import (
	"context"
	"errors"
	"testing"

	"github.com/sawthorn/merch-wallet/internal/repository"
	"github.com/stretchr/testify/require"
)

type mockUserRepo struct {
	getUser       func(context.Context, string) (*repository.User, error)
	changeBalance func(context.Context, int, int) error
}

func (r *mockUserRepo) GetUserByUsername(ctx context.Context, username string) (*repository.User, error) {
	return r.getUser(ctx, username)
}

func (r *mockUserRepo) ChangeBalance(ctx context.Context, userID int, delta int) error {
	return r.changeBalance(ctx, userID, delta)
}

type mockTransferRepo struct {
	createTransfer func(context.Context, int, int, int) error
}

func (r *mockTransferRepo) CreateTransfer(ctx context.Context, from int, to int, amount int) error {
	return r.createTransfer(ctx, from, to, amount)
}

type mockTxManager struct{}

func (m *mockTxManager) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func TestTransfer_Success(t *testing.T) {
	userA := repository.NewUser(
		repository.WithUsername("A"),
		repository.WithID(1),
		repository.WithBalance(1000),
	)
	userB := repository.NewUser(
		repository.WithUsername("B"),
		repository.WithID(2),
		repository.WithBalance(1000),
	)

	transferAmount := 10

	ts := NewTransferService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				if username == userA.Username() {
					return userA, nil
				}
				if username == userB.Username() {
					return userB, nil
				}
				t.Fatalf("getting non-existent user: %s", username)
				return nil, repository.ErrUserNotFound
			},
			changeBalance: func(ctx context.Context, userID, delta int) error {
				if userID == userA.ID() && delta == -transferAmount {
					return nil
				}
				if userID == userB.ID() && delta == transferAmount {
					return nil
				}
				t.Fatalf("unexpected balance change: ID %d by %d", userID, delta)
				return errors.New("")
			},
		},
		&mockTransferRepo{
			createTransfer: func(ctx context.Context, from, to, amount int) error {
				if from != userA.ID() {
					t.Fatalf("sender ID incorrect. got: %d, want: %d", from, userA.ID())
				}
				if to != userB.ID() {
					t.Fatalf("reciever ID incorrect. got: %d, want:%d", to, userB.ID())
				}
				if amount != transferAmount {
					t.Fatalf("transfer amont incorrect. got: %d, want: %d", amount, transferAmount)
				}
				return nil
			},
		},
		&mockTxManager{},
	)

	err := ts.TransferCoinsToUser(
		context.Background(),
		userA.Username(),
		userB.Username(),
		transferAmount,
	)
	require.NoError(t, err)
}

func TestTransfer_InsufficientBalance(t *testing.T) {
	userA := repository.NewUser(
		repository.WithUsername("A"),
		repository.WithID(1),
		repository.WithBalance(1),
	)
	userB := repository.NewUser(
		repository.WithUsername("B"),
		repository.WithID(2),
		repository.WithBalance(1000),
	)

	transferAmount := 10000

	ts := NewTransferService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				if username == userA.Username() {
					return userA, nil
				}
				if username == userB.Username() {
					return userB, nil
				}
				t.Fatalf("getting non-existent user: %s", username)
				return nil, repository.ErrUserNotFound
			},
			changeBalance: func(ctx context.Context, userID, delta int) error {
				t.Fatal("changeBalance must not run in this test")
				return nil
			},
		},
		&mockTransferRepo{
			createTransfer: func(ctx context.Context, from, to, amount int) error {
				t.Fatal("createTransfer must not run in this test")
				return nil
			},
		},
		&mockTxManager{},
	)

	err := ts.TransferCoinsToUser(
		context.Background(),
		userA.Username(),
		userB.Username(),
		transferAmount,
	)
	require.ErrorIs(t, err, ErrInsufficientBalance)
}

func TestTransfer_RecipientNotFound(t *testing.T) {
	userA := repository.NewUser(
		repository.WithUsername("A"),
		repository.WithID(1),
		repository.WithBalance(1000),
	)

	transferAmount := 10

	ts := NewTransferService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				if username == userA.Username() {
					return userA, nil
				}
				return nil, repository.ErrUserNotFound
			},
			changeBalance: func(ctx context.Context, userID, delta int) error {
				t.Fatal("changeBalance must not run in this test")
				return nil
			},
		},
		&mockTransferRepo{
			createTransfer: func(ctx context.Context, from, to, amount int) error {
				t.Fatal("createTransfer must not run in this test")
				return nil
			},
		},
		&mockTxManager{},
	)

	err := ts.TransferCoinsToUser(
		context.Background(),
		userA.Username(),
		"B",
		transferAmount,
	)
	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestTransfer_TransferToSelfRestricted(t *testing.T) {
	userA := repository.NewUser(
		repository.WithUsername("A"),
		repository.WithID(1),
		repository.WithBalance(1000),
	)

	transferAmount := 10

	ts := NewTransferService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				t.Fatal("getUser must not run in this test")
				return nil, nil
			},
			changeBalance: func(ctx context.Context, userID, delta int) error {
				t.Fatal("changeBalance must not run in this test")
				return nil
			},
		},
		&mockTransferRepo{
			createTransfer: func(ctx context.Context, from, to, amount int) error {
				t.Fatal("createTransfer must not run in this test")
				return nil
			},
		},
		&mockTxManager{},
	)

	err := ts.TransferCoinsToUser(
		context.Background(),
		userA.Username(),
		userA.Username(),
		transferAmount,
	)
	require.ErrorIs(t, err, ErrTransferToSelfProhibited)
}

func TestTransfer_InvalidTransferAmount(t *testing.T) {
	transferAmount := -10

	ts := NewTransferService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				t.Fatal("getUser must not run in this test")
				return nil, nil
			},
			changeBalance: func(ctx context.Context, userID, delta int) error {
				t.Fatal("changeBalance must not run in this test")
				return nil
			},
		},
		&mockTransferRepo{
			createTransfer: func(ctx context.Context, from, to, amount int) error {
				t.Fatal("createTransfer must not run in this test")
				return nil
			},
		},
		&mockTxManager{},
	)

	err := ts.TransferCoinsToUser(
		context.Background(),
		"A",
		"B",
		transferAmount,
	)
	require.ErrorIs(t, err, ErrInvalidTransferAmount)
}

func TestTransfer_ContextCanceled(t *testing.T) {
	ts := NewTransferService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				t.Fatal("getUser must not run in this test, context must be canceled")
				return nil, nil
			},
			changeBalance: func(ctx context.Context, userID, delta int) error {
				t.Fatal("changeBalance must not run in this test")
				return nil
			},
		},
		&mockTransferRepo{
			createTransfer: func(ctx context.Context, from, to, amount int) error {
				t.Fatal("createTransfer must not run in this test")
				return nil
			},
		},
		&mockTxManager{},
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := ts.TransferCoinsToUser(
		ctx,
		"A",
		"B",
		1,
	)
	require.ErrorIs(t, err, context.Canceled)
}
