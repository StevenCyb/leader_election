package internal

// OrderedList is a list of numbers that is always sorted.
type OrderedList[T Number] []T

// AddOrdered adds a new value to the SortedList, keeping it sorted.
func (o *OrderedList[T]) AddOrdered(uid T) {
	i := 0

	for i < len(*o) && (*o)[i] < uid {
		i++
	}

	*o = append((*o)[:i], append([]T{uid}, (*o)[i:]...)...)
}

// RemoveOrdered removes a value from the SortedList, keeping it sorted.
func (o *OrderedList[T]) RemoveOrdered(uid T) {
	index := o.GetIndexFor(uid)
	if index != -1 {
		*o = append((*o)[:index], (*o)[index+1:]...)
	}
}

// Get returns the value at the given index in the SortedList.
func (o OrderedList[T]) GetIndexFor(number T) int {
	for i, v := range o {
		if v == number {
			return i
		}
	}

	return -1
}

func (o OrderedList[T]) GetValueForIndexLooped(index int) T {
	return o[index%len(o)]
}

func (o OrderedList[T]) GetValueForIndexLoopedReverted(index int) T {
	if index >= len(o) {
		index = len(o) - ((index % len(o)) + 2)
	}

	return o[index]
}

func (o OrderedList[T]) GetIndexLeftOfValue(number T) *int {
	for i, v := range o {
		if v == number {
			prevIndex := (i - 1 + len(o)) % len(o)

			return &prevIndex
		}
	}

	return nil
}

func (o OrderedList[T]) GetIndexRightOfValue(number T) *int {
	for i, v := range o {
		if v == number {
			nextIndex := (i + 1) % len(o)

			return &nextIndex
		}
	}

	return nil
}
