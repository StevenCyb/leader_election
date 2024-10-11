package internal

import "sync"

// SafeValue is a generic struct that protects access to a value with a mutex.
type SafeValue[T any] struct {
	value T
	mutex *sync.RWMutex
}

// NewSafeValue creates a new SafeValue with an initial value.
func NewSafeValue[T any](initial T) *SafeValue[T] {
	return &SafeValue[T]{
		value: initial,
		mutex: &sync.RWMutex{},
	}
}

// NewSafeValue creates a new SafeValue with an initial value.
func NewSafeValueWithSharedMutex[T any](initial T, mutex *sync.RWMutex) *SafeValue[T] {
	return &SafeValue[T]{
		value: initial,
		mutex: mutex,
	}
}

// Set safely sets the value.
func (s *SafeValue[T]) Set(val T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.value = val
}

// Get safely retrieves the value.
func (s *SafeValue[T]) Get() T {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.value
}

// Define SortedList as a slice of uint64.
type OrderedUIDList []uint64

// AddOrdered adds a new value to the SortedList, keeping it sorted.
func (o *OrderedUIDList) AddOrdered(uid uint64) {
	i := 0

	for i < len(*o) && (*o)[i] > uid {
		i++
	}

	*o = append((*o)[:i], append([]uint64{uid}, (*o)[i:]...)...)
}

// GetNext returns the next value in the SortedList, wrapping around.
func (o OrderedUIDList) GetNext(index int) uint64 {
	return o[index%len(o)]
}

// FindNeighbor finds the index of the next value to the given number.
func (o OrderedUIDList) FindNeighbor(number uint64) *int {
	for i, v := range o {
		if v == number {
			nextIndex := (i + 1) % len(o)

			return &nextIndex
		}
	}

	return nil
}
