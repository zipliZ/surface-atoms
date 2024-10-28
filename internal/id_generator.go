package internal

type IdGenerator struct {
	counter int
}

func NewIdGenerator() *IdGenerator {
	return &IdGenerator{counter: 1}
}

func (i *IdGenerator) GetId() int {
	id := i.counter
	i.counter++
	return id
}
