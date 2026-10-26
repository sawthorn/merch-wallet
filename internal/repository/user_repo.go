package repository

import (
	"github.com/sawthorn/merch-wallet/internal/repository/transactions"
)

type UserRepository struct {
	executor *transactions.DBExecutor
}

func NewUserRepository(executor *transactions.DBExecutor) *UserRepository {
	return &UserRepository{
		executor: executor,
	}
}
