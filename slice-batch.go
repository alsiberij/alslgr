package alslgr

type (
	SliceBatch[T any] []T
)

var (
	_ Batch[[]any, any] = (*SliceBatch[any])(nil)
)

func (s *SliceBatch[T]) IsFull() bool {
	return len(*s) == cap(*s)
}

func (s *SliceBatch[T]) Append(data T) {
	*s = append(*s, data)
}

func (s *SliceBatch[T]) Extract() []T {
	cp := make([]T, len(*s))
	copy(cp, *s)
	*s = (*s)[:0]
	return cp
}
