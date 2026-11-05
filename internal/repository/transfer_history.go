package repository

import "time"

type TransferHistoryEntry struct {
	username  string
	amount    int
	createdAt time.Time
}

func (th *TransferHistoryEntry) Username() string {
	return th.username
}

func (th *TransferHistoryEntry) Amount() int {
	return th.amount
}

func (th *TransferHistoryEntry) CreatedAt() time.Time {
	return th.createdAt
}
