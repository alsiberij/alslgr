package alslgr

type (
	SliceBatchProducer[T any] int
)

var (
	_ BatchProducer[[]any, any] = (*SliceBatchProducer[any])(nil)
)

func (s *SliceBatchProducer[T]) NewBatch() Batch[[]T, T] {
	batch := make(SliceBatch[T], 0, *s)
	return &batch
}
