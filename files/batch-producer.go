package files

import "github.com/alsiberij/alslgr/v3"

type (
	DataBatchProducer struct {
		batchMaxLen int
	}
)

var (
	_ alslgr.DataBatchProducer[[][]byte, []byte] = (*DataBatchProducer)(nil)
)

func NewDataBatchProducer(batchMaxLen int) DataBatchProducer {
	return DataBatchProducer{
		batchMaxLen: batchMaxLen,
	}
}

func (b *DataBatchProducer) NewDataBatch() alslgr.DataBatch[[][]byte, []byte] {
	return &DataBatch{
		aggregatedData: make([][]byte, 0, b.batchMaxLen),
	}
}
