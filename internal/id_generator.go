package internal

// IdGenerator генератор уникальных идентификаторов.
type IdGenerator struct {
	counter int
}

func NewIdGenerator() *IdGenerator {
	return &IdGenerator{counter: 1}
}

// GetId возвращает текущий идентификатор и увеличивает счётчик.
func (i *IdGenerator) GetId() int {
	id := i.counter
	i.counter++
	return id
}
