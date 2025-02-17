package internal

// IdGenerator is a unique identifier generator.
type IdGenerator struct {
	counter int
}

// NewIdGenerator creates a new ID generator instance.
func NewIdGenerator() *IdGenerator {
	return &IdGenerator{counter: 1}
}

// GetId returns the current identifier and increments the counter.
func (i *IdGenerator) GetId() int {
	id := i.counter
	i.counter++
	return id
}
