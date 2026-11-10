package repository

type Item struct {
	id   int
	name string
	cost int
}

func NewItem(id int, name string, cost int) *Item {
	return &Item{id: id, name: name, cost: cost}
}

func (i *Item) ID() int {
	return i.id
}

func (i *Item) Name() string {
	return i.name
}

func (i *Item) Cost() int {
	return i.cost
}
