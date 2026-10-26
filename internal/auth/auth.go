// Package auth implements auth service
// for user registration and authentication.
package auth

import (
	"context"
	"errors"

	"github.com/sawthorn/merch-wallet/internal/repository"
	"github.com/sawthorn/merch-wallet/internal/repository/transactions"
	"golang.org/x/crypto/bcrypt"
)

var ErrPasswordMismatch = errors.New("passwords do not match")

type AuthService struct {
	userRepo  *repository.UserRepository
	trManager *transactions.TransactionManager
}

func NewAuthService(
	userRepo *repository.UserRepository,
	trManager *transactions.TransactionManager,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		trManager: trManager,
	}
}

func (as *AuthService) LoginOrRegisterUser(ctx context.Context, username, password string) (string, error) {
	var user *repository.User
	var err error
	user, err = as.getOrCreateUser(ctx, username, password)
	if err != nil {
		return "", err
	}
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password()),
		[]byte(password),
	)
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return "", ErrPasswordMismatch
	}
	if err != nil {
		return "", err
	}

	return createToken(username)
}

// getOrCreateUser retrieves user by username or creates that user and returns it
// it has retry logic to handle user creation race between two requests
func (as *AuthService) getOrCreateUser(ctx context.Context, username, password string) (*repository.User, error) {
	var user *repository.User
	var err error

	err = as.trManager.RunInTx(ctx, func(ctx context.Context) error {
		user, err = as.userRepo.GetUserByUsername(ctx, username)
		if err == nil {
			return nil
		}
		// db err
		if !errors.Is(err, repository.ErrUserNotFound) {
			return err
		}

		user, err = as.userRepo.CreateUser(ctx, username, password)
		if err == nil {
			return nil
		}
		if !errors.Is(err, repository.ErrUserAlreadyExists) {
			return err
		}

		// another request created user first
		user, err = as.userRepo.GetUserByUsername(ctx, username)
		return err
	},
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}
