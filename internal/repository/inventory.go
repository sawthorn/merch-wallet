package repository

type InventoryItem struct {
	name     string
	quantity int
}

func NewInventoryItem(name string, quantity uint) InventoryItem {
	return InventoryItem{name: name, quantity: int(quantity)}
}

func (i *InventoryItem) Name() string {
	return i.name
}

func (i *InventoryItem) Quantity() int {
	return i.quantity
}
