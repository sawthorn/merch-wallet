package info

import (
	"context"
	"errors"
	"time"

	"github.com/sawthorn/merch-wallet/internal/repository"
)

var ErrUserNotFound = errors.New("user not found")

type (
	userRepo interface {
		GetUserByUsername(ctx context.Context, username string) (*repository.User, error)
	}
	inventoryRepo interface {
		GetUserInventory(ctx context.Context, userID int) ([]repository.InventoryItem, error)
	}
	transferRepo interface {
		GetSentHistory(ctx context.Context, userID int) ([]repository.TransferHistoryEntry, error)
		GetReceivedHistory(ctx context.Context, userID int) ([]repository.TransferHistoryEntry, error)
	}
)

type InfoService struct {
	userRepo      userRepo
	inventoryRepo inventoryRepo
	transferRepo  transferRepo
}

func NewInfoService(
	userRepo userRepo,
	inventoryRepo inventoryRepo,
	transferRepo transferRepo,
) *InfoService {
	return &InfoService{
		userRepo:      userRepo,
		inventoryRepo: inventoryRepo,
		transferRepo:  transferRepo,
	}
}

type UserInfo struct {
	Balance          int
	Inventory        []InventoryItem
	TransferSent     []TransferHistoryEntry
	TransferRecieved []TransferHistoryEntry
}

type InventoryItem struct {
	Name     string
	Quantity int
}

type TransferHistoryEntry struct {
	Username  string
	Amount    int
	CreatedAt time.Time
}

func (is *InfoService) CollectUserInformation(ctx context.Context, username string) (*UserInfo, error) {
	userInfo := &UserInfo{}
	user, err := is.userRepo.GetUserByUsername(ctx, username)
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	userInfo.Balance = user.Balance()

	inv, err := is.inventoryRepo.GetUserInventory(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	for _, item := range inv {
		userInfo.Inventory = append(userInfo.Inventory, InventoryItem{
			Name:     item.Name(),
			Quantity: item.Quantity(),
		})
	}

	sent, err := is.transferRepo.GetSentHistory(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	recieved, err := is.transferRepo.GetReceivedHistory(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	for _, tr := range sent {
		userInfo.TransferSent = append(userInfo.TransferSent, TransferHistoryEntry{
			Username:  tr.Username(),
			Amount:    tr.Amount(),
			CreatedAt: tr.CreatedAt(),
		})
	}

	for _, tr := range recieved {
		userInfo.TransferRecieved = append(userInfo.TransferRecieved, TransferHistoryEntry{
			Username:  tr.Username(),
			Amount:    tr.Amount(),
			CreatedAt: tr.CreatedAt(),
		})
	}

	return userInfo, nil
}
