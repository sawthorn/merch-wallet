package repository

import (
	"errors"
	"time"
)

type Transfer struct {
	fromID    int
	toID      int
	amount    int
	createdAt time.Time
}

var (
	ErrInvalidAmount         = errors.New("invalid transfer amount")
	ErrSameSenderAndReceiver = errors.New("same sender and receiver")
)

func NewTransfer(fromID int, toID int, amount int) (*Transfer, error) {
	if fromID == toID {
		return nil, ErrSameSenderAndReceiver
	}

	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	return &Transfer{
		fromID: fromID,
		toID:   toID,
		amount: amount,
	}, nil
}

func (t *Transfer) FromID() int {
	return t.fromID
}

func (t *Transfer) ToID() int {
	return t.toID
}

func (t *Transfer) Amount() int {
	return t.amount
}

func (t *Transfer) CreatedAt() time.Time {
	return t.createdAt
}
