package info

import (
	"context"
	"testing"

	"github.com/sawthorn/merch-wallet/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserRepo struct {
	getUser func(ctx context.Context, username string) (*repository.User, error)
}

func (r *mockUserRepo) GetUserByUsername(ctx context.Context, username string) (*repository.User, error) {
	return r.getUser(ctx, username)
}

type mockInventoryRepo struct {
	getInventory func(ctx context.Context, userID int) ([]repository.InventoryItem, error)
}

func (r *mockInventoryRepo) GetUserInventory(ctx context.Context, userID int) ([]repository.InventoryItem, error) {
	return r.getInventory(ctx, userID)
}

type mockTransferRepo struct {
	getSentHistory     func(ctx context.Context, userID int) ([]repository.TransferHistoryEntry, error)
	getReceivedHistory func(ctx context.Context, userID int) ([]repository.TransferHistoryEntry, error)
}

func (r *mockTransferRepo) GetSentHistory(ctx context.Context, userID int) ([]repository.TransferHistoryEntry, error) {
	return r.getSentHistory(ctx, userID)
}

func (r *mockTransferRepo) GetReceivedHistory(ctx context.Context, userID int) ([]repository.TransferHistoryEntry, error) {
	return r.getReceivedHistory(ctx, userID)
}

type mockTxManager struct{}

func (m *mockTxManager) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func toRepoItems(items []InventoryItem) []repository.InventoryItem {
	res := make([]repository.InventoryItem, 0, len(items))
	for _, v := range items {
		res = append(res, repository.NewInventoryItem(v.Name, uint(v.Quantity)))
	}
	return res
}

func toRepoTransferHistory(history []TransferHistoryEntry) []repository.TransferHistoryEntry {
	res := make([]repository.TransferHistoryEntry, 0, len(history))
	for _, v := range history {
		res = append(res, repository.NewTransferHistoryEntry(v.Username, uint(v.Amount)))
	}
	return res
}

func TestCollectUserInfo_Success(t *testing.T) {
	user := repository.NewUser(
		repository.WithUsername("A"),
		repository.WithBalance(1000),
	)

	inventory := []repository.InventoryItem{
		repository.NewInventoryItem("ItemA", 1),
		repository.NewInventoryItem("ItemB", 2),
	}

	sent := []repository.TransferHistoryEntry{
		repository.NewTransferHistoryEntry("B", 10),
		repository.NewTransferHistoryEntry("C", 100),
	}

	recieved := []repository.TransferHistoryEntry{
		repository.NewTransferHistoryEntry("D", 10),
		repository.NewTransferHistoryEntry("E", 100),
	}

	is := NewInfoService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				if username == user.Username() {
					return user, nil
				}
				t.Errorf("getting unexpected user: %s", username)
				return nil, ErrUserNotFound
			},
		},
		&mockInventoryRepo{
			getInventory: func(ctx context.Context, userID int) ([]repository.InventoryItem, error) {
				if userID == user.ID() {
					return inventory, nil
				}
				t.Errorf("getInventory for unexpected user: %d", userID)
				return nil, nil
			},
		},
		&mockTransferRepo{
			getSentHistory: func(ctx context.Context, userID int) ([]repository.TransferHistoryEntry, error) {
				if userID == user.ID() {
					return sent, nil
				}
				t.Errorf("getSentHistory for unexpected user: %d", userID)
				return nil, nil
			},
			getReceivedHistory: func(ctx context.Context, userID int) ([]repository.TransferHistoryEntry, error) {
				if userID == user.ID() {
					return recieved, nil
				}
				t.Errorf("getReceivedHistory for unexpected user: %d", userID)
				return nil, nil
			},
		},
	)

	userInfo, err := is.CollectUserInformation(
		context.Background(),
		user.Username(),
	)
	require.NoError(t, err)
	assert.Equal(t, userInfo.Balance, user.Balance())
	assert.Equal(t, toRepoItems(userInfo.Inventory), inventory)
	assert.Equal(t, toRepoTransferHistory(userInfo.TransferSent), sent)
	assert.Equal(t, toRepoTransferHistory(userInfo.TransferRecieved), recieved)
}

func TestCollectUserInfo_UserNotFound(t *testing.T) {
	is := NewInfoService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				return nil, ErrUserNotFound
			},
		},
		&mockInventoryRepo{
			getInventory: func(ctx context.Context, userID int) ([]repository.InventoryItem, error) {
				t.Error("getInventory mustn't be called in this test")
				return nil, nil
			},
		},
		&mockTransferRepo{
			getSentHistory: func(ctx context.Context, userID int) ([]repository.TransferHistoryEntry, error) {
				t.Error("getSentHistory mustn't be called in this test")
				return nil, nil
			},
			getReceivedHistory: func(ctx context.Context, userID int) ([]repository.TransferHistoryEntry, error) {
				t.Error("getReceivedHistory mustn't be called in this test")
				return nil, nil
			},
		},
	)

	userInfo, err := is.CollectUserInformation(
		context.Background(),
		"A",
	)
	require.ErrorIs(t, err, ErrUserNotFound)
	require.Nil(t, userInfo)
}

func TestCollectUserInfo_InfoEmpty(t *testing.T) {
	user := repository.NewUser(
		repository.WithUsername("A"),
		repository.WithBalance(1000),
	)
	inventory := []repository.InventoryItem{}
	sent := []repository.TransferHistoryEntry{}
	recieved := []repository.TransferHistoryEntry{}

	is := NewInfoService(
		&mockUserRepo{
			getUser: func(ctx context.Context, username string) (*repository.User, error) {
				if username == user.Username() {
					return user, nil
				}
				t.Errorf("getting unexpected user: %s", username)
				return nil, ErrUserNotFound
			},
		},
		&mockInventoryRepo{
			getInventory: func(ctx context.Context, userID int) ([]repository.InventoryItem, error) {
				if userID == user.ID() {
					return inventory, nil
				}
				t.Errorf("getInventory for unexpected user: %d", userID)
				return nil, nil
			},
		},
		&mockTransferRepo{
			getSentHistory: func(ctx context.Context, userID int) ([]repository.TransferHistoryEntry, error) {
				if userID == user.ID() {
					return sent, nil
				}
				t.Errorf("getSentHistory for unexpected user: %d", userID)
				return nil, nil
			},
			getReceivedHistory: func(ctx context.Context, userID int) ([]repository.TransferHistoryEntry, error) {
				if userID == user.ID() {
					return recieved, nil
				}
				t.Errorf("getReceivedHistory for unexpected user: %d", userID)
				return nil, nil
			},
		},
	)

	userInfo, err := is.CollectUserInformation(
		context.Background(),
		user.Username(),
	)
	require.NoError(t, err)
	assert.Equal(t, userInfo.Balance, user.Balance())
	assert.Empty(t, userInfo.Inventory)
	assert.Empty(t, userInfo.TransferSent)
	assert.Empty(t, userInfo.TransferRecieved)
}
