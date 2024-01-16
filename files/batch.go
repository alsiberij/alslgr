package files

import (
	"github.com/alsiberij/alslgr/v3"
)

type (
	Batch struct {
		aggregatedData [][]byte
	}
)

var (
	_ alslgr.Batch[[][]byte, []byte] = (*Batch)(nil)
)

func (b *Batch) ReadyToSend() bool {
	return len(b.aggregatedData) == cap(b.aggregatedData)
}

func (b *Batch) Append(data []byte) {
	b.aggregatedData = append(b.aggregatedData, data)
}

func (b *Batch) Extract() [][]byte {
	return b.aggregatedData
}
