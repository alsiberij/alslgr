package file

import (
	"github.com/alsiberij/alslgr/v3"
)

type (
	// batchProducer implements alslgr.BatchProducer
	batchProducer int
)

// NewBatch returns new batch with defined limited capacity
func (p *batchProducer) NewBatch() alslgr.Batch[[]byte, []byte] {
	b := make(batch, 0, *p)
	return &b
}
