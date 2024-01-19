package file

import (
	"strings"
	"sync"
	"testing"
)

// FOR CORRECT RESULTS ADD -benchtime=1x FLAG

const (
	// Amount of concurrent writers that will try to write data in forwarder
	benchGoroutines = 1_000

	// Constant is used for creating large data entries
	benchStringsRepeat = 20

	// Amount of data that will be aggregated into a batch
	benchBatchSize = 200

	// Maximum length of underlying buffer that will be used before writing data batch into a file
	// Otherwise all entries will be written consequentially
	benchMaxBufferLen = benchBatchSize * 1_000

	// Size of buffers of internal channels
	benchChannelBuffer = 32

	benchFilename1 = "test-bench-1.txt"
	benchFilename2 = "test-bench-2.txt"
)

func BenchmarkFileBufferedForwarder(b *testing.B) {
	bfwd := NewBufferedForwarder(Config{
		BatchMaxLen:           benchBatchSize,
		MaxForwarderBufferLen: benchMaxBufferLen,
		Filename:              benchFilename1,
		LastResortWriter:      nil, // In case of error writing file panic will occur
		ChannelsBuffer:        benchChannelBuffer,
	})
	var wg sync.WaitGroup

	b.ResetTimer()

	wg.Add(benchGoroutines)
	for j := 0; j < benchGoroutines; j++ {
		go func() {
			bfwd.Write([]byte(strings.Repeat("HELLO WORLD FROM HERE LONG TEXT STARTS RIGHT HERE\n", benchStringsRepeat)))
			wg.Done()
		}()
	}
	wg.Wait()
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

	wg.Add(benchGoroutines)
	for j := 0; j < benchGoroutines; j++ {
		go func() {
			fwd.Forward([]byte(strings.Repeat("HELLO WORLD FROM HERE LONG TEXT STARTS RIGHT HERE\n", benchStringsRepeat)))
			wg.Done()
		}()
	}
	wg.Wait()
}
