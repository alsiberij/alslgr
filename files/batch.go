package files

import (
	"github.com/alsiberij/alslgr/v3"
)

type (
	DataBatch struct {
		aggregatedData [][]byte
	}
)

var (
	_ alslgr.DataBatch[[][]byte, []byte] = (*DataBatch)(nil)
)

func (b *DataBatch) ReadyToSend() bool {
	return len(b.aggregatedData) == cap(b.aggregatedData)
}

func (b *DataBatch) Append(data []byte) {
	b.aggregatedData = append(b.aggregatedData, data)
}

func (b *DataBatch) Extract() [][]byte {
	return b.aggregatedData
}
