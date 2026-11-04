package purchase

import (
	"context"
	"errors"

	"github.com/sawthorn/merch-wallet/internal/repository"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrItemNotFound        = repository.ErrItemNotFound
)

type userRepo interface {
	GetUserByUsername(ctx context.Context, username string) (*repository.User, error)
	ChangeBalance(ctx context.Context, userID int, delta int) error
}

type itemRepo interface {
	GetItemByName(ctx context.Context, itemName string) (*repository.Item, error)
}

type purchaseRepo interface {
	CreatePurchase(ctx context.Context, userID int, itemID int) error
}

type inventoryRepo interface {
	AddItemToUserInventory(ctx context.Context, userID int, itemID int) error
}

type transactionManager interface {
	RunInTx(context.Context, func(context.Context) error) error
}

type PurchaseService struct {
	userRepo      userRepo
	itemRepo      itemRepo
	purchaseRepo  purchaseRepo
	inventoryRepo inventoryRepo
	trManager     transactionManager
}

func NewPurchaseService(
	ur userRepo,
	itemr itemRepo,
	pr purchaseRepo,
	invr inventoryRepo,
	tr transactionManager,
) *PurchaseService {
	return &PurchaseService{
		userRepo:      ur,
		itemRepo:      itemr,
		purchaseRepo:  pr,
		inventoryRepo: invr,
		trManager:     tr,
	}
}

func (ps *PurchaseService) BuyItem(ctx context.Context, username string, itemName string) error {
	return ps.trManager.RunInTx(ctx, func(ctx context.Context) error {
		user, err := ps.userRepo.GetUserByUsername(ctx, username)
		if err != nil {
			return err
		}

		item, err := ps.itemRepo.GetItemByName(ctx, itemName)
		if errors.Is(err, repository.ErrItemNotFound) {
			return ErrItemNotFound
		}
		if err != nil {
			return err
		}

		if user.Balance() < item.Cost() {
			return ErrInsufficientBalance
		}

		err = ps.userRepo.ChangeBalance(ctx, user.ID(), -item.Cost())
		if errors.Is(err, repository.ErrInsufficientBalance) {
			return ErrInsufficientBalance
		}
		if err != nil {
			return err
		}

		err = ps.inventoryRepo.AddItemToUserInventory(ctx, user.ID(), item.ID())
		if err != nil {
			return err
		}

		return ps.purchaseRepo.CreatePurchase(ctx, user.ID(), item.ID())
	})
}
