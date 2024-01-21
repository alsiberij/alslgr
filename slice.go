package alslgr

type (
	// Slice is an example of the implementation Batch interface. This instance uses regular slice for batching data
	// and considered full when capacity of underlying slice equals its length. Calling the Extract method will
	// return copy of the underlying slice and reset the length of the original one to zero
	Slice[T any] []T
)

var (
	_ Batch[[]any, any] = (*Slice[any])(nil)
)

// IsFull returns true when length of underlying slice equals its capacity
func (s *Slice[T]) IsFull() bool {
	return len(*s) == cap(*s)
}

// Append appends data to the underlying slice
func (s *Slice[T]) Append(data T) {
	*s = append(*s, data)
}

// Extract return copy of underlying slice and reset length to zero of the original one for reuse
func (s *Slice[T]) Extract() []T {
	cp := make([]T, len(*s))
	copy(cp, *s)
	*s = (*s)[:0]
	return cp
}

type (
	// SliceProducer can be considered as a Slice factory which returns a new slice on every NewBatch
	// call.
	SliceProducer[T any] int
)

var (
	_ BatchProducer[[]any, any] = (*SliceProducer[any])(nil)
)

// NewBatch will make a new Slice with zero length and defined capacity
func (s *SliceProducer[T]) NewBatch() Batch[[]T, T] {
	batch := make(Slice[T], 0, *s)
	return &batch
}