package file

import (
	"github.com/alsiberij/alslgr/v3"
)

type (
	batchProducer int
)

func (p *batchProducer) NewBatch() alslgr.Batch[[]byte, []byte] {
	b := make(batch, 0, *p)
	return &b
}
