package stdout

import (
	"github.com/alsiberij/alslgr/v3"
	"os"
	"sync"
)

type (
	DataForwarder struct {
		mu *sync.Mutex
	}
)

var (
	_ alslgr.DataForwarder[[][]byte, []byte] = (*DataForwarder)(nil)
)

func NewDataForwarder() DataForwarder {
	return DataForwarder{
		mu: &sync.Mutex{},
	}
}

func (f *DataForwarder) Reopen() {
	// NO OP
}

func (f *DataForwarder) ForwardDataBatch(batch [][]byte) {
	for _, data := range batch {
		f.mu.Lock()
		_, _ = os.Stdout.Write(data)
		f.mu.Unlock()
	}
}

func (f *DataForwarder) ForwardData(data []byte) {
	f.mu.Lock()
	_, _ = os.Stdout.Write(data)
	f.mu.Unlock()
}

func (f *DataForwarder) Close() {
	// NO OP
}
