package alslgr

import (
	"context"
	"sync"
	"time"
)

type (
	// Logger is a buffered storage for writing logs. It is required to call Start before any Write.
	// It is not allowed to copy instance of Logger after calling Start
	Logger struct {
		capacity   int
		baseBuf    []byte
		baseLen    int
		reserveBuf []byte
		reserveLen int

		dumpCh  chan []byte
		dumpMu  sync.Mutex
		writeMu sync.Mutex

		dumper Dumper

		autoDumpCancelFunc context.CancelFunc
	}
)

// NewLogger returns Logger with defined capacity and dumper,
// where all data will be dumped on buffer overflow or by some timeout.
func NewLogger(capacity int, dumper Dumper) Logger {
	return Logger{
		capacity:   capacity,
		baseBuf:    make([]byte, capacity),
		reserveBuf: make([]byte, capacity),
		dumpCh:     make(chan []byte),
		dumper:     dumper,
	}
}

// Start is necessary to call in order to start dumping goroutine.
func (l *Logger) Start() {
	go l.dumpWorker()
}

// Stop stops auto dump goroutine, perform manual dump of internal buffer
// It is not allowed to call ant Write after Stop
func (l *Logger) Stop() {
	if l.autoDumpCancelFunc != nil {
		l.autoDumpCancelFunc()
	}
	l.ManualDump()
	close(l.dumpCh)
}

// Write writes b to internal buffer.
// If there is no space in it, new buffer will be available immediately.
// Old buffer will be asynchronously dumped with internal Dumper
func (l *Logger) Write(b []byte) {
	bLen := len(b)

	l.writeMu.Lock()
	defer l.writeMu.Unlock()

	if bLen > l.capacity-l.baseLen {
		l.dumpMu.Lock()

		l.swapBuffers()

		l.dumpCh <- l.reserveBuf[:l.reserveLen]

		if bLen > l.capacity {
			l.dumpMu.Lock()
			l.dumpCh <- b
			return
		}
	}

	l.baseLen += copy(l.baseBuf[l.baseLen:], b)

	return
}

// ManualDump performs manual blocking dump of internal buffer.
func (l *Logger) ManualDump() {
	l.writeMu.Lock()
	l.dumpMu.Lock()

	l.swapBuffers()

	if l.reserveLen > 0 {
		l.dumper.Dump(l.reserveBuf[:l.reserveLen])
		l.reserveLen = 0
	}

	l.dumpMu.Unlock()
	l.writeMu.Unlock()
}

// StartAutoDump starts async goroutine that will Dump internal buffer
func (l *Logger) StartAutoDump(interval time.Duration) {
	if l.autoDumpCancelFunc == nil {
		ctx, cancel := context.WithCancel(context.Background())
		l.autoDumpCancelFunc = cancel
		go l.autoDump(ctx, interval)
	}
}

func (l *Logger) autoDump(ctx context.Context, interval time.Duration) {
	for {
		select {
		case <-ctx.Done():
		case <-time.After(interval):
			l.writeMu.Lock()
			l.dumpMu.Lock()

			l.swapBuffers()

			l.dumpCh <- l.reserveBuf[:l.reserveLen]

			l.writeMu.Unlock()
		}
	}
}

func (l *Logger) dumpWorker() {
	for data := range l.dumpCh {
		if len(data) > 0 {
			l.dumper.Dump(data)
			l.reserveLen = 0
		}
		l.dumpMu.Unlock()
	}
}

func (l *Logger) swapBuffers() {
	l.baseBuf, l.reserveBuf = l.reserveBuf, l.baseBuf
	l.baseLen, l.reserveLen = l.reserveLen, l.baseLen
}
