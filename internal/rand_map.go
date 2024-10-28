package internal

import (
	"crypto/rand"
	"math/big"
	"sync"
)

type RandMap[K comparable, V any] struct {
	m sync.RWMutex

	// Where the objects you care about are stored.
	container map[K]V

	// A slice of the map keys used in the map above. We put them in a slice
	// so that we can get a random key by choosing a random index.
	keys []K

	// We store the index of each key, so that when we remove an item, we can
	// quickly remove it from the slice above.
	sliceKeyIndex map[K]int
}

func NewRandMap[K comparable, V any]() *RandMap[K, V] {
	return &RandMap[K, V]{
		container:     make(map[K]V),
		sliceKeyIndex: make(map[K]int),
	}
}

func (s *RandMap[K, V]) Add(key K, item V) {
	s.m.Lock()
	defer s.m.Unlock()

	// store object in map
	s.container[key] = item

	// add map key to slice of map keys
	s.keys = append(s.keys, key)

	// store the index of the map key
	index := len(s.keys) - 1
	s.sliceKeyIndex[key] = index
}

func (s *RandMap[K, V]) Get(key K) (val V, ok bool) {
	s.m.RLock()
	defer s.m.RUnlock()

	item, ok := s.container[key]
	if !ok {
		return *new(V), false
	}

	return item, true
}

func (s *RandMap[K, V]) Remove(key K) {
	s.m.Lock()
	defer s.m.Unlock()

	s.remove(key)
}

// caller is responsible for locking
func (s *RandMap[K, V]) remove(key K) {
	// get index in key slice for key
	index, exists := s.sliceKeyIndex[key]
	if !exists {
		// item does not exist
		return
	}

	delete(s.sliceKeyIndex, key)

	wasLastIndex := len(s.keys)-1 == index

	// remove key from slice of keys
	s.keys[index] = s.keys[len(s.keys)-1]
	s.keys = s.keys[:len(s.keys)-1]

	// we just swapped the last element to another position.
	// so we need to update it's index (if it was not in last position)
	if !wasLastIndex {
		otherKey := s.keys[index]
		s.sliceKeyIndex[otherKey] = index
	}

	// remove object from map
	delete(s.container, key)
}

func (s *RandMap[K, V]) Random() (key K, val V, ok bool) {
	s.m.RLock()
	defer s.m.RUnlock()

	if len(s.keys) == 0 {
		return *new(K), *new(V), false
	}

	randomIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(s.keys))))
	key = s.keys[(randomIndex).Int64()]

	item := s.container[key]

	return key, item, true
}

func (s *RandMap[K, V]) PopRandom() (val V, ok bool) {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.container) == 0 {
		return *new(V), false
	}

	randomIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(s.keys))))
	key := s.keys[(randomIndex).Int64()]

	item, ok := s.container[key]
	if !ok {
		return *new(V), false
	}

	s.remove(key)

	return item, true
}

func (s *RandMap[K, V]) Len() int {
	s.m.RLock()
	defer s.m.RUnlock()

	return len(s.container)
}
