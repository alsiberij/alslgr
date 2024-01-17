package files

import "github.com/alsiberij/alslgr/v3"

type (
	BatchProducer struct {
		batchMaxLen int
	}
)

var (
	_ alslgr.DataBatchProducer[[][]byte, []byte] = (*BatchProducer)(nil)
)

func NewDataBatchProducer(batchMaxLen int) BatchProducer {
	return BatchProducer{
		batchMaxLen: batchMaxLen,
	}
}

func (b *BatchProducer) NewDataBatch() alslgr.DataBatch[[][]byte, []byte] {
	return &Batch{
		aggregatedData: make([][]byte, 0, b.batchMaxLen),
	}
}
