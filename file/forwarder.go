package file

import (
	"bytes"
	"github.com/alsiberij/alslgr/v3"
	"io"
	"os"
	"sync"
)

type (
	forwarder struct {
		filename string

		mu               *sync.Mutex
		writerCloser     io.WriteCloser
		lastResortWriter io.Writer

		maxBufferLen int
	}
)

var (
	_ alslgr.Writer[[][]byte, []byte] = (*forwarder)(nil)
)

func newForwarder(filename string, lastResortWriter io.Writer, maxBufferLen int) forwarder {
	var w io.WriteCloser

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		w = file
	}

	f := forwarder{
		mu:               &sync.Mutex{},
		writerCloser:     w,
		lastResortWriter: lastResortWriter,
		maxBufferLen:     maxBufferLen,
	}

	return f
}

func (f *forwarder) Reset() {
	f.mu.Lock()

	if f.writerCloser != nil {
		_ = f.writerCloser.Close()
	}

	file, err := os.OpenFile(f.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		f.writerCloser = nil
	} else {
		f.writerCloser = file
	}

	f.mu.Unlock()
}

func (f *forwarder) WriteBatch(batch [][]byte) {
	var size int
	for _, data := range batch {
		size += len(data)
	}

	if size == 0 {
		return
	}

	if size > f.maxBufferLen {
		for _, b := range batch {
			f.Write(b)
		}
		return
	}

	buf := bytes.NewBuffer(make([]byte, 0, size))

	for _, data := range batch {
		_, _ = buf.Write(data)
	}

	f.Write(buf.Bytes())
}

func (f *forwarder) Write(data []byte) {
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

func (f *forwarder) Close() {
	f.mu.Lock()
	if f.writerCloser != nil {
		_ = f.writerCloser.Close()
	}
	f.mu.Unlock()
}
