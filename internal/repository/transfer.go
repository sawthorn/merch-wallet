package repository

type Transfer struct {
	fromID int
	toID   int
	amount int
}

func NewTransfer(fromID int, toID int, amount int) *Transfer {
	return &Transfer{
		fromID: fromID,
		toID:   toID,
		amount: amount,
	}
}
