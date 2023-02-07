package alslgr

import (
	"context"
	"sync"
	"time"
)

type (
	logger struct {
		mx sync.Mutex

		buffer   []byte
		length   int
		capacity int

		dumper Dumper
	}
)

func NewLogger(capacity int, dumper Dumper) Logger {
	return &logger{
		mx:       sync.Mutex{},
		buffer:   make([]byte, capacity),
		capacity: capacity,
		dumper:   dumper,
	}
}

func (l *logger) Write(b []byte) (int, error) {
	l.mx.Lock()
	defer l.mx.Unlock()

	return len(b), l.write(b)
}

func (l *logger) write(b []byte) error {
	bLen := len(b)

	if bLen > l.capacity {
		return l.dumper.Dump(b)
	}

	if l.capacity-l.length < bLen {
		err := l.dump()
		if err != nil {
			return err
		}
	}

	copy(l.buffer[l.length:], b)
	l.length += bLen

	return nil
}

func (l *logger) DumpBuffer() error {
	l.mx.Lock()
	defer l.mx.Unlock()

	return l.dump()
}

func (l *logger) dump() error {
	if l.length == 0 {
		return nil
	}

	err := l.dumper.Dump(l.buffer[:l.length])

	if err == nil {
		l.length = 0
	}

	return err
}

func (l *logger) AutoDumpBuffer(interval time.Duration) (<-chan error, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go repeatOpWorker(ctx, interval, errCh, l.DumpBuffer)

	return errCh, cancel
}
