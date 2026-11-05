package repository

type InventoryItem struct {
	name     string
	quantity int
}

func (i *InventoryItem) Name() string {
	return i.name
}

func (i *InventoryItem) Quantity() int {
	return i.quantity
}
