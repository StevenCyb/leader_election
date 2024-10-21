package internal

type Tuple[T, U any] struct {
	First  T
	Second U
}

func NewTuple[T, U any](first T, second U) Tuple[T, U] {
	return Tuple[T, U]{First: first, Second: second}
}
