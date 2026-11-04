package repository

type Item struct {
	id   int
	name string
	cost int
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
