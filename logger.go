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

// NewLogger returns a Logger interface which is an abstraction of logger having internal buffer with defined capacity
// and implemented dumper. It is recommended to define pretty big capacity in order to minimize Dumper calls.
// In other words, if average length of log entry equals 100 bytes and capacity is 10000 bytes, then Dumper calls amount
// will decrease in 100 times. But keep in mind that dumping operation is blocking and in case of VERY big capacity
// slow Dumper call will block other write operations. You can get rid of it by implementing non-blocking Dumper.
func NewLogger(capacity int, dumper Dumper) Logger {
	return &logger{
		mx:       sync.Mutex{},
		buffer:   make([]byte, capacity),
		capacity: capacity,
		dumper:   dumper,
	}
}

func (l *logger) Write(b []byte) (int, error) {
	bLen := len(b)

	l.mx.Lock()
	defer l.mx.Unlock()

	if l.capacity-l.length < bLen {
		err := l.dumper.Dump(l.buffer[:l.length])
		if err != nil {
			return 0, err
		}

		l.length = 0

		if l.capacity < bLen {
			err = l.dumper.Dump(b)
			if err != nil {
				return 0, err
			}
			return bLen, nil
		}
	}

	l.length += copy(l.buffer[l.length:], b)

	return bLen, nil
}

func (l *logger) DumpBuffer() error {
	l.mx.Lock()
	defer l.mx.Unlock()

	err := l.dumper.Dump(l.buffer[:l.length])
	if err == nil {
		l.length = 0
	}

	return err
}

func (l *logger) AutoDumpBuffer(interval time.Duration) (<-chan error, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)

	go func(ctx context.Context, interval time.Duration, errCh chan<- error, operation func() error) {
		defer close(errCh)

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				errCh <- operation()
			}
		}
	}(ctx, interval, errCh, l.DumpBuffer)

	return errCh, cancel
}
