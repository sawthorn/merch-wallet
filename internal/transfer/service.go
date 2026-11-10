package transfer

import (
	"context"
	"errors"
	"fmt"

	"github.com/sawthorn/merch-wallet/internal/repository"
)

var (
	ErrInsufficientBalance      = errors.New("not enough balance to perform this operation")
	ErrRecipientNotFound        = errors.New("recipient not found")
	ErrTransferToSelfProhibited = errors.New("transfer to self prohibited")
	ErrInvalidTransferAmount    = errors.New("invalid transfer amount")
)

type userRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*repository.User, error)
	ChangeBalance(ctx context.Context, userID int, delta int) error
}

type transferRepository interface {
	CreateTransfer(ctx context.Context, from int, to int, amount int) error
}

type transactionManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type TransferService struct {
	userRepo     userRepository
	transferRepo transferRepository
	trManager    transactionManager
}

func NewTransferService(
	userRepo userRepository,
	transferRepo transferRepository,
	trManager transactionManager,
) *TransferService {
	return &TransferService{
		userRepo:     userRepo,
		transferRepo: transferRepo,
		trManager:    trManager,
	}
}

func (ts *TransferService) TransferCoinsToUser(ctx context.Context, fromUsername string, toUsername string, amount int) error {
	return ts.trManager.RunInTx(ctx, func(ctx context.Context) error {
		if amount <= 0 {
			return ErrInvalidTransferAmount
		}
		if fromUsername == toUsername {
			return ErrTransferToSelfProhibited
		}
		fromUser, err := ts.userRepo.GetUserByUsername(ctx, fromUsername)
		if err != nil {
			return err
		}

		if fromUser.Balance() < amount {
			return ErrInsufficientBalance
		}

		if errors.Is(ctx.Err(), context.Canceled) {
			return nil
		}

		toUser, err := ts.userRepo.GetUserByUsername(ctx, toUsername)
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrRecipientNotFound
		}
		if err != nil {
			return err
		}

		err = ts.userRepo.ChangeBalance(ctx, fromUser.ID(), -amount)
		if errors.Is(err, repository.ErrInsufficientBalance) {
			return ErrInsufficientBalance
		}
		if err != nil {
			return fmt.Errorf("sender DecreaseBalance failed: %w", err)
		}
		err = ts.userRepo.ChangeBalance(ctx, toUser.ID(), amount)
		if err != nil {
			return fmt.Errorf("reciever IncreaseBalance failed: %w", err)
		}

		err = ts.transferRepo.CreateTransfer(ctx, fromUser.ID(), toUser.ID(), amount)
		if err != nil {
			return fmt.Errorf("create transfer record failed: %w", err)
		}

		return nil
	})
}
