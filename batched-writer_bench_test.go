package alslgr

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	// Indicates how much sending batch of data is more efficient than sending that amount of data without batching
	// Value of 5 means that sending batch of 5 values has the same speed as sending single value
	// More effective batch sending means more effectivity for forwarding at all
	batchEfficiencyCoefficient = 5

	// Amount of concurrent writers that will try to write data in writer
	goroutines = 300

	// Amount of data that will be aggregated into a batch
	batchSize = 200

	// Amount of time which takes writing a single value in destination writer
	writeDelay = time.Millisecond * 3

	// Size of buffers of internal channels
	channelBuffer = 32
)

type (
	// BenchForwarder is an abstract data writer, which takes d time to write any data and requires exclusive access
	// Field i is used to check that correct amount of data was written
	BenchForwarder struct {
		d  time.Duration
		mu sync.Mutex
		i  int
	}
)

func (b *BenchForwarder) Reset() {
}

func (b *BenchForwarder) WriteBatch(batch [][]byte) {
	b.mu.Lock()
	time.Sleep(b.d * (time.Duration)(len(batch)/batchEfficiencyCoefficient+1))
	b.i += len(batch)
	b.mu.Unlock()
}

func (b *BenchForwarder) Write(_ []byte) {
	b.mu.Lock()
	time.Sleep(b.d)
	b.i++
	b.mu.Unlock()
}

func (b *BenchForwarder) Close() {
}

func BenchmarkBufferedForwarder(b *testing.B) {
	// Default slice batching is used here
	sbp := SliceProducer[[]byte](batchSize)

	fwd := BenchForwarder{
		d:  writeDelay,
		mu: sync.Mutex{},
	}

	bfwd := NewBatchedWriter(Config[[][]byte, []byte]{
		BatchProducer:   &sbp,
		Writer:          &fwd,
		ChannelsBuffer:  channelBuffer,
		BatchingWorkers: runtime.NumCPU()/2 + 1,
		WritingWorkers:  runtime.NumCPU()/2 + 1,
	})

	var wg sync.WaitGroup

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(goroutines)
		for j := 0; j < goroutines; j++ {
			go func() {
				bfwd.Write(nil)
				wg.Done()
			}()
		}
		wg.Wait()
	}

	bfwd.Close()

	if fwd.i != goroutines*b.N {
		b.Fatal("MISMATCH")
	}
}

func BenchmarkForwarder(b *testing.B) {
	var wg sync.WaitGroup
	fwd := BenchForwarder{
		d:  writeDelay,
		mu: sync.Mutex{},
	}

	for i := 0; i < b.N; i++ {
		wg.Add(goroutines)
		for j := 0; j < goroutines; j++ {
			go func() {
				fwd.Write(nil)
				wg.Done()
			}()
		}
		wg.Wait()
	}

	if fwd.i != goroutines*b.N {
		b.Fatal("MISMATCH")
	}
}
