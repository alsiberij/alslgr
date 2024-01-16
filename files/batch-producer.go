package files

import "github.com/alsiberij/alslgr/v3"

type (
	BatchProducer struct {
		batchMaxLen int
	}
)

var (
	_ alslgr.BatchProducer[[][]byte, []byte] = (*BatchProducer)(nil)
)

func NewBatchProducer(batchMaxLen int) BatchProducer {
	return BatchProducer{
		batchMaxLen: batchMaxLen,
	}
}

func (b *BatchProducer) NewBatch() alslgr.Batch[[][]byte, []byte] {
	return &Batch{
		aggregatedData: make([][]byte, 0, b.batchMaxLen),
	}
}
