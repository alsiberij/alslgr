package file

import (
	"github.com/alsiberij/alslgr/v3"
	"io"
	"os"
	"sync"
)

type (
	// writer implements alslgr.Writer with []byte for B and T and writes data in internal file.
	// If any error occurs, lastResortWriter will be used. Calling Reset will reopen same file.
	writer struct {
		filename string

		mu               *sync.Mutex
		writerCloser     io.WriteCloser
		lastResortWriter io.Writer
	}
)

var (
	_ alslgr.Writer[[]byte, []byte] = (*writer)(nil)
)

func newWriter(filename string, lastResortWriter io.Writer) (writer, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0644)
	if err != nil {
		return writer{}, err
	}

	return writer{
		filename:         filename,
		mu:               &sync.Mutex{},
		writerCloser:     file,
		lastResortWriter: lastResortWriter,
	}, nil
}

func (f *writer) Reset() {
	f.mu.Lock()

	if f.writerCloser != nil {
		_ = f.writerCloser.Close()
	}

	file, err := os.OpenFile(f.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0644)
	if err != nil {
		f.writerCloser = nil
	} else {
		f.writerCloser = file
	}

	f.mu.Unlock()
}

func (f *writer) WriteBatch(batch []byte) {
	f.Write(batch)
}

func (f *writer) Write(data []byte) {
	f.mu.Lock()

	var writeSucceed bool
	if f.writerCloser != nil {
		n, err := f.writerCloser.Write(data)
		if err == nil && n == len(data) {
			writeSucceed = true
		} else {
			data = data[n:]
		}
	}

	if !writeSucceed {
		_, _ = f.lastResortWriter.Write(data)
	}

	f.mu.Unlock()
}

func (f *writer) Close() {
	f.mu.Lock()
	if f.writerCloser != nil {
		_ = f.writerCloser.Close()
	}
	f.mu.Unlock()
}
