package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sawthorn/merch-wallet/internal/repository/transactions"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInsufficientBalance = errors.New("insufficient balance")
	PasswordHashingCost    = 15
)

type UserRepository struct {
	executor *transactions.DBExecutor
}

func NewUserRepository(executor *transactions.DBExecutor) *UserRepository {
	return &UserRepository{
		executor: executor,
	}
}

func (ur *UserRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	tx := ur.executor.For(ctx)
	user := &User{}
	err := tx.QueryRowContext(ctx, `
		SELECT id, username, password, balance
		FROM users
		WHERE username=$1
		`, username).Scan(
		&user.id,
		&user.username,
		&user.password,
		&user.balance,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (ur *UserRepository) CreateUser(ctx context.Context, username, password string) (*User, error) {
	tx := ur.executor.For(ctx)
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), PasswordHashingCost)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	user := &User{}
	err = tx.QueryRowContext(ctx, `
		INSERT INTO users(username, password, balance)
		VALUES($1, $2, 1000)
		ON CONFLICT (username) DO NOTHING
		RETURNING username, password, balance
		`, username, string(hashed)).Scan(
		&user.username,
		&user.password,
		&user.balance,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("user insert failed: %w", err)
	}

	return user, nil
}

func (ur *UserRepository) ChangeBalance(ctx context.Context, userID int, delta int) error {
	tx := ur.executor.For(ctx)
	res, err := tx.ExecContext(ctx, `
		UPDATE users
		SET balance = balance + $1
		WHERE id = $2 AND balance + $1 >= 0
		`, delta, userID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrInsufficientBalance
	}

	return nil
}
