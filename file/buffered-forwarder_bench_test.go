package file

import (
	"strings"
	"sync"
	"testing"
)

// IT IS RECOMMENDED TO ADD -benchtime=10s FLAG

const (
	// benchGoroutines amount of concurrent writers that will try to write data in forwarder
	benchGoroutines = 1000

	// benchString constant is a data that wil be written in file
	benchString = "HELLO WORLD FROM HERE LONG TEXT STARTS RIGHT HERE\n"
	// benchStringsRepeat constant is used for creating large data entries
	benchStringsRepeat = 10

	// benchBatchSize is amount of data that will be aggregated into a batch
	benchBatchSize = 100

	// benchMaxBufferLen is maximum length of underlying buffer that will be used before writing data batch into a file
	// Otherwise all entries will be written consequentially
	benchMaxBufferLen = len(benchString) * benchStringsRepeat * benchBatchSize

	// benchChannelBuffer is a size of buffers of internal channels
	benchChannelBuffer = 32

	benchFilename1 = "test-bench-buffered-forwarder.txt"
	benchFilename2 = "test-bench-forwarder.txt"
)

func BenchmarkFileBufferedForwarder(b *testing.B) {
	bfwd := NewBufferedForwarder(Config{
		BatchMaxLen:               benchBatchSize,
		MaxBufferLenBeforeWriting: benchMaxBufferLen,
		Filename:                  benchFilename1,
		LastResortWriter:          nil, // In case of error writing file panic will occur
		ChannelsBuffer:            benchChannelBuffer,
	})
	var wg sync.WaitGroup

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(benchGoroutines)
		for j := 0; j < benchGoroutines; j++ {
			go func() {
				bfwd.Write([]byte(strings.Repeat(benchString, benchStringsRepeat)))
				wg.Done()
			}()
		}
		wg.Wait()
	}

	bfwd.Close()
}

func BenchmarkFileForwarder(b *testing.B) {
	fwd := newForwarder(
		benchFilename2,
		nil, // In case of error writing file panic will occur
		benchMaxBufferLen,
	)

	var wg sync.WaitGroup

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(benchGoroutines)
		for j := 0; j < benchGoroutines; j++ {
			go func() {
				fwd.Write([]byte(strings.Repeat(benchString, benchStringsRepeat)))
				wg.Done()
			}()
		}
		wg.Wait()
	}
}
