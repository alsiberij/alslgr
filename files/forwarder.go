package files

import (
	"bytes"
	"github.com/alsiberij/alslgr/v3"
	"io"
	"os"
	"sync"
)

type (
	Forwarder struct {
		filename string

		writerErrorsHandler WriterErrorsHandler

		mu               *sync.Mutex
		writerCloser     io.WriteCloser
		lastResortWriter io.WriteCloser

		maxBufferLen int
	}
)

var (
	_ alslgr.DataForwarder[[][]byte, []byte] = (*Forwarder)(nil)
)

func NewForwarder(filename string, writerErrorsHandler WriterErrorsHandler,
	lastResortWriter io.WriteCloser, maxBufferLen int) (Forwarder, error) {

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return Forwarder{}, err
	}

	f := Forwarder{
		writerErrorsHandler: writerErrorsHandler,
		mu:                  &sync.Mutex{},
		writerCloser:        file,
		lastResortWriter:    lastResortWriter,
		maxBufferLen:        maxBufferLen,
	}

	return f, err
}

func (f *Forwarder) Reopen() {
	f.mu.Lock()

	_ = f.writerCloser.Close()

	file, err := os.OpenFile(f.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		f.writerErrorsHandler.HandleError(f.lastResortWriter, nil, err)
		f.writerCloser = f.lastResortWriter
	} else {
		f.writerCloser = file
	}

	f.mu.Unlock()
}

func (f *Forwarder) ForwardBatch(batch [][]byte) {
	var size int
	for _, data := range batch {
		size += len(data)
	}

	if size == 0 {
		return
	}

	if size > f.maxBufferLen {
		for _, b := range batch {
			f.Forward(b)
		}
		return
	}

	buf := bytes.NewBuffer(make([]byte, 0, size))

	for _, data := range batch {
		_, _ = buf.Write(data)
	}

	f.Forward(buf.Bytes())
}

func (f *Forwarder) Forward(data []byte) {
	f.mu.Lock()
	_, err := f.writerCloser.Write(data)
	f.mu.Unlock()

	if err != nil {
		f.writerErrorsHandler.HandleError(f.lastResortWriter, data, err)
	}
}

func (f *Forwarder) Close() {
	f.mu.Lock()
	_ = f.writerCloser.Close()
	f.mu.Unlock()
}
