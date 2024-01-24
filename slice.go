package alslgr

type (
	// Slice is an example of the implementation of Batch interface. This instance uses regular slice for batching data
	// and considered full when capacity of the underlying slice equals its length. Calling the Extract method will
	// return a copy of the underlying slice and reset the length of the original one to zero
	Slice[T any] []T
)

var (
	_ Batch[[]any, any] = (*Slice[any])(nil)
)

// TryAppend appends data to the underlying slice and returns true when the length of the underlying slice is
// less its capacity, false otherwise
func (s *Slice[T]) TryAppend(data T) bool {
	if len(*s) == cap(*s) {
		return false
	}
	*s = append(*s, data)
	return true
}

// Extract returns a copy of the underlying slice and resets the length of the original one to zero for further reuse
func (s *Slice[T]) Extract() []T {
	cp := make([]T, len(*s))
	copy(cp, *s)
	*s = (*s)[:0]
	return cp
}

// Pack packs data into a new slice
func (s *Slice[T]) Pack(data T) []T {
	return []T{data}
}

type (
	// SliceProducer is an example of the implementation of SliceProducer interface. It can be considered as
	// a Slice factory that returns a new slice on every NewBatch call with capacity of the underlying value
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
