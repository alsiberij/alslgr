package alslgr

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// FOR CORRECT RESULTS ADD -benchtime=1x FLAG

const (
	// Indicates how much sending batch of data is more efficient than sending that amount of data without batching
	// Value of 3 means that sending batch of 3 values has the same speed as sending single value
	// More effective batch sending means more effectivity for forwarding at all
	batchEfficiencyCoefficient = 3

	// Amount of concurrent writers that will try to write data in forwarder
	goroutines = 1000

	// Amount of data that will be aggregated into a batch
	batchSize = 200

	// Amount of time which takes writing a single value in destination writer
	writeDelay = time.Millisecond * 3

	// Size of buffers of internal channels
	channelBuffer = 1
)

type (
	// BenchForwarder is an abstract data forwarder, which takes d time to write any data and requires exclusive access
	// Field i is used to check that correct amount of data was written
	BenchForwarder struct {
		d  time.Duration
		mu sync.Mutex
		i  int
	}
)

func (b *BenchForwarder) Reset() {
}

func (b *BenchForwarder) ForwardBatch(batch [][]byte) {
	b.mu.Lock()
	time.Sleep(b.d * (time.Duration)(len(batch)/batchEfficiencyCoefficient+1))
	b.i += len(batch)
	b.mu.Unlock()
}

func (b *BenchForwarder) Forward(_ []byte) {
	b.mu.Lock()
	time.Sleep(b.d)
	b.i++
	b.mu.Unlock()
}

func (b *BenchForwarder) Close() {
}

func BenchmarkBufferedForwarder(b *testing.B) {
	// Default slice batching is used here
	sbp := SliceBatchProducer[[]byte](batchSize)
	fwd := BenchForwarder{
		d:  writeDelay,
		mu: sync.Mutex{},
	}
	bfwd := NewBufferedForwarder(Config[[][]byte, []byte]{
		BatchProducer:         &sbp,
		Forwarder:             &fwd,
		ChannelsBuffer:        channelBuffer,
		BatchingConcurrency:   runtime.NumCPU(),
		ForwardingConcurrency: runtime.NumCPU(),
	})
	var wg sync.WaitGroup

	b.ResetTimer()

	wg.Add(goroutines)
	for j := 0; j < goroutines; j++ {
		go func() {
			bfwd.Write(nil)
			wg.Done()
		}()
	}
	wg.Wait()
	bfwd.Close()
	if fwd.i != goroutines {
		b.Fatal("MISMATCH")
	}
}

func BenchmarkForwarder(b *testing.B) {
	var wg sync.WaitGroup
	fwd := BenchForwarder{
		d:  writeDelay,
		mu: sync.Mutex{},
	}

	wg.Add(goroutines)
	for j := 0; j < goroutines; j++ {
		go func() {
			fwd.Forward(nil)
			wg.Done()
		}()
	}
	wg.Wait()
	if fwd.i != goroutines {
		b.Fatal("MISMATCH")
	}
}
