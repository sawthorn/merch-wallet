package purchase

import (
	"context"
	"testing"

	"github.com/sawthorn/merch-wallet/internal/repository"
	"github.com/stretchr/testify/require"
)

type mockUserRepo struct {
	getUser       func(ctx context.Context, username string) (*repository.User, error)
	changeBalance func(ctx context.Context, userID int, delta int) error
}

func (r *mockUserRepo) GetUserByUsername(ctx context.Context, username string) (*repository.User, error) {
	return r.getUser(ctx, username)
}

func (r *mockUserRepo) ChangeBalance(ctx context.Context, userID int, delta int) error {
	return r.changeBalance(ctx, userID, delta)
}

type mockItemRepo struct {
	getItem func(ctx context.Context, itemName string) (*repository.Item, error)
}

func (r *mockItemRepo) GetItemByName(ctx context.Context, itemName string) (*repository.Item, error) {
	return r.getItem(ctx, itemName)
}

type mockPurchaseRepo struct {
	createPurchase func(ctx context.Context, userID int, itemID int) error
}

func (r *mockPurchaseRepo) CreatePurchase(ctx context.Context, userID int, itemID int) error {
	return r.createPurchase(ctx, userID, itemID)
}

type mockInventoryRepo struct {
	addItem func(ctx context.Context, userID int, itemID int) error
}

func (r *mockInventoryRepo) AddItemToUserInventory(ctx context.Context, userID int, itemID int) error {
	return r.addItem(ctx, userID, itemID)
}

type mockTxManager struct{}

func (m *mockTxManager) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func TestBuyItem_Success(t *testing.T) {
	user := repository.NewUser(
		repository.WithUsername("TestUser"),
		repository.WithID(1),
		repository.WithBalance(1000),
	)
	item := repository.NewItem(1, "TestItem", 10)

	ps := NewPurchaseService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				if username == user.Username() {
					return user, nil
				}
				return nil, repository.ErrUserNotFound
			},
			changeBalance: func(ctx context.Context, userID, delta int) error {
				if userID != user.ID() {
					t.Fatalf("wrong user ID. want: %d, got: %d", user.ID(), userID)
					return nil
				}
				if delta != -item.Cost() {
					t.Fatalf("wrong balance change for purchase. want: %d, got: %d", -item.Cost(), delta)
				}
				return nil
			},
		},
		&mockItemRepo{
			getItem: func(ctx context.Context, itemName string) (*repository.Item, error) {
				if itemName == item.Name() {
					return item, nil
				}
				return nil, repository.ErrItemNotFound
			},
		},
		&mockPurchaseRepo{
			createPurchase: func(ctx context.Context, userID, itemID int) error {
				if userID != user.ID() {
					t.Fatalf("wrong user ID. want: %d, got: %d", user.ID(), userID)
				}
				if itemID != item.ID() {
					t.Fatalf("wrong item ID. want: %d, got: %d", item.ID(), itemID)
				}
				return nil
			},
		},
		&mockInventoryRepo{
			addItem: func(ctx context.Context, userID, itemID int) error {
				if userID != user.ID() {
					t.Fatalf("wrong user ID. want: %d, got: %d", user.ID(), userID)
				}
				if itemID != item.ID() {
					t.Fatalf("wrong item ID. want: %d, got: %d", item.ID(), itemID)
				}
				return nil
			},
		},
		&mockTxManager{},
	)

	err := ps.BuyItem(
		context.Background(),
		user.Username(),
		item.Name(),
	)
	require.NoError(t, err)
}

func TestBuyItem_ItemNotFound(t *testing.T) {
	user := repository.NewUser(
		repository.WithUsername("TestUser"),
		repository.WithID(1),
		repository.WithBalance(1000),
	)

	ps := NewPurchaseService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				if username == user.Username() {
					return user, nil
				}
				return nil, repository.ErrUserNotFound
			},
			changeBalance: func(ctx context.Context, userID, delta int) error {
				t.Fatal("changeBalance mustn't be called in this test")
				return nil
			},
		},
		&mockItemRepo{
			getItem: func(ctx context.Context, itemName string) (*repository.Item, error) {
				return nil, repository.ErrItemNotFound
			},
		},
		&mockPurchaseRepo{
			createPurchase: func(ctx context.Context, userID, itemID int) error {
				t.Fatal("createPurchase mustn't be called in this test")
				return nil
			},
		},
		&mockInventoryRepo{
			addItem: func(ctx context.Context, userID, itemID int) error {
				t.Fatal("addItem mustn't be called in this test")
				return nil
			},
		},
		&mockTxManager{},
	)

	err := ps.BuyItem(
		context.Background(),
		user.Username(),
		"NotExists",
	)
	require.ErrorIs(t, err, ErrItemNotFound)
}

func TestBuyItem_InsufficientBalance(t *testing.T) {
	user := repository.NewUser(
		repository.WithUsername("TestUser"),
		repository.WithID(1),
		repository.WithBalance(10),
	)
	item := repository.NewItem(1, "TestItem", 1000)

	ps := NewPurchaseService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				if username == user.Username() {
					return user, nil
				}
				return nil, repository.ErrUserNotFound
			},
			changeBalance: func(ctx context.Context, userID, delta int) error {
				if userID != user.ID() {
					t.Fatalf("wrong user ID. want: %d, got: %d", user.ID(), userID)
					return nil
				}
				if delta != -item.Cost() {
					t.Fatalf("wrong balance change for purchase. want: %d, got: %d", -item.Cost(), delta)
				}

				if user.Balance() < item.Cost() {
					return ErrInsufficientBalance
				}
				return nil
			},
		},
		&mockItemRepo{
			getItem: func(ctx context.Context, itemName string) (*repository.Item, error) {
				if itemName == item.Name() {
					return item, nil
				}
				return nil, repository.ErrItemNotFound
			},
		},
		&mockPurchaseRepo{
			createPurchase: func(ctx context.Context, userID, itemID int) error {
				t.Fatal("createPurchase mustn't be called in this test")
				return nil
			},
		},
		&mockInventoryRepo{
			addItem: func(ctx context.Context, userID, itemID int) error {
				t.Fatal("addItem mustn't be called in this test")
				return nil
			},
		},
		&mockTxManager{},
	)

	err := ps.BuyItem(
		context.Background(),
		user.Username(),
		item.Name(),
	)
	require.ErrorIs(t, err, ErrInsufficientBalance)
}

func TestBuyItem_ContextCanceled(t *testing.T) {
	user := repository.NewUser(
		repository.WithUsername("TestUser"),
		repository.WithID(1),
		repository.WithBalance(10),
	)
	item := repository.NewItem(1, "TestItem", 1000)

	ps := NewPurchaseService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				t.Fatal("getUser must not run in this test, context must be canceled")
				return nil, nil
			},
			changeBalance: func(ctx context.Context, userID, delta int) error {
				if userID != user.ID() {
					t.Fatalf("wrong user ID. want: %d, got: %d", user.ID(), userID)
					return nil
				}
				if delta != -item.Cost() {
					t.Fatalf("wrong balance change for purchase. want: %d, got: %d", -item.Cost(), delta)
				}

				if user.Balance() < item.Cost() {
					return ErrInsufficientBalance
				}
				return nil
			},
		},
		&mockItemRepo{
			getItem: func(ctx context.Context, itemName string) (*repository.Item, error) {
				t.Fatal("createPurchase mustn't be called in this test")
				return nil, nil
			},
		},
		&mockPurchaseRepo{
			createPurchase: func(ctx context.Context, userID, itemID int) error {
				t.Fatal("createPurchase mustn't be called in this test")
				return nil
			},
		},
		&mockInventoryRepo{
			addItem: func(ctx context.Context, userID, itemID int) error {
				t.Fatal("addItem mustn't be called in this test")
				return nil
			},
		},
		&mockTxManager{},
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := ps.BuyItem(
		ctx,
		user.Username(),
		item.Name(),
	)
	require.ErrorIs(t, err, context.Canceled)
}
