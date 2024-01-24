package alslgr

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

const (
	// batchEfficiency shows much is more efficient to send batch of data than doing it one by one
	batchEfficiency = 100

	// writerSleepTime is an imitation of I/O delay per one data entry
	writerSleepTime = time.Millisecond * 3

	// writerPoolSize is a size of imitated connection pool. In this test Writer doesn't require exclusive access,
	// so we can use multiple workers (Number of workers should be <= writerPoolSize)
	writerPoolSize = 16

	// batchMaxSize is a maximum size of every accumulated batch
	batchMaxSize = 100

	// batchedWriterWorkers is a number of workers used by BatchedWriter
	batchedWriterWorkers = 16

	// goroutinesWriting is a number of goroutines that will write data in BatchedWriter
	goroutinesWriting = 100

	// dataRepeatPerGoroutineWriting is a number of repeats of data inside every writing goroutine
	dataRepeatPerGoroutineWriting = 1000

	// savingBatchesTickerInterval is the time between writing accumulated batches whether that are ready or not
	savingBatchesTickerInterval = time.Second / 10
)

type (
	writer struct {
		poolCh        chan *int64
		resetIsCalled bool
	}
)

func (w *writer) Reset() {
	w.resetIsCalled = true
}

func (w *writer) WriteBatch(batch [][]byte) {
	connection := <-w.poolCh
	time.Sleep(writerSleepTime * time.Duration(len(batch)/batchEfficiency+1))
	*connection += int64(len(batch))
	w.poolCh <- connection
}

func (w *writer) Write(_ []byte) {
	connection := <-w.poolCh
	time.Sleep(writerSleepTime)
	*connection++
	w.poolCh <- connection
}

func (w *writer) Close() {
	close(w.poolCh)
}

// Sum returns number of written data entries across all pool connections to verify data integrity
func (w *writer) Sum() int64 {
	var sum int64
	for val := range w.poolCh {
		sum += *val
	}
	return sum
}

func TestBatchedWriter(t *testing.T) {
	// Initializing imitated connections pool
	w := writer{
		poolCh: make(chan *int64, writerPoolSize),
	}
	for i := 0; i < writerPoolSize; i++ {
		w.poolCh <- new(int64)
	}

	// Initialing slice batch producer
	bp := SliceProducer[[]byte](batchMaxSize)

	// Initializing BatchedWriter
	bw := NewBatchedWriter[[][]byte, []byte](&bp, &w, batchedWriterWorkers)

	wg := sync.WaitGroup{}
	// This atomic is required to correct amount of written data in case where context was cancelled before writing
	var contextWriteCorrection atomic.Int64

	for i := 0; i < goroutinesWriting; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < dataRepeatPerGoroutineWriting/2; j++ {
				// Regular write
				bw.Write(nil)

				// Context write
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Now().UnixNano()%750)*time.Microsecond)
				err := bw.WriteCtx(ctx, nil)
				if err != nil {
					contextWriteCorrection.Add(1)
				}
				cancel()
			}
			wg.Done()
		}()
	}

	bw.ResetWriter()
	bw.SaveBatches()

	wg.Wait()

	bw.Close()
	w.Close()

	sum := w.Sum()
	expected := goroutinesWriting*(dataRepeatPerGoroutineWriting/2)*2 - contextWriteCorrection.Load()

	if sum != expected {
		t.Fatalf("Expected %d, got %d", expected, sum)
	}
}

func TestBatchedWriterTicker(t *testing.T) {
	// Initializing imitated connections pool
	w := writer{
		poolCh: make(chan *int64, writerPoolSize),
	}
	for i := 0; i < writerPoolSize; i++ {
		w.poolCh <- new(int64)
	}

	// Initialing slice batch producer
	bp := SliceProducer[[]byte](batchMaxSize)

	// Initializing BatchedWriter
	bw := NewBatchedWriter[[][]byte, []byte](&bp, &w, batchedWriterWorkers)

	defer close(AutoWritingBatches(&bw, savingBatchesTickerInterval))

	wg := sync.WaitGroup{}

	for i := 0; i < goroutinesWriting; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < dataRepeatPerGoroutineWriting; j++ {
				bw.Write(nil)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	bw.Close()
	w.Close()

	sum := w.Sum()
	expected := int64(goroutinesWriting * dataRepeatPerGoroutineWriting)

	if sum != expected {
		t.Fatalf("Expected %d, got %d", expected, sum)
	}
}

func TestBatchedWriterSigHup(t *testing.T) {
	sigCh := make(chan os.Signal)

	// Initializing imitated connections pool
	w := writer{
		poolCh: make(chan *int64, writerPoolSize),
	}
	for i := 0; i < writerPoolSize; i++ {
		w.poolCh <- new(int64)
	}

	// Initialing slice batch producer
	bp := SliceProducer[[]byte](batchMaxSize)

	// Initializing BatchedWriter
	bw := NewBatchedWriter[[][]byte, []byte](&bp, &w, batchedWriterWorkers)

	defer close(ResetWriterOnSigHup(&bw, sigCh))

	wg := sync.WaitGroup{}

	for i := 0; i < goroutinesWriting; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < dataRepeatPerGoroutineWriting; j++ {
				bw.Write(nil)
			}
			wg.Done()
		}()
	}

	go func() {
		sigCh <- syscall.SIGHUP
	}()

	wg.Wait()

	bw.Close()
	w.Close()

	if !w.resetIsCalled {
		t.Fatalf("Reset was not called")
	}
}
