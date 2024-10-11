package internal

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
