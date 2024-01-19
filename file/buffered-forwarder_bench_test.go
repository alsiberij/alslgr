package file

import (
	"strings"
	"sync"
	"testing"
)

// FOR CORRECT RESULTS ADD -benchtime=1x FLAG

const (
	// Amount of concurrent writers that will try to write data in forwarder
	goroutines = 1_000

	// Constant is used for creating large data entries
	stringsRepeat = 20

	// Amount of data that will be aggregated into a batch
	batchSize = 200

	// Maximum length of underlying buffer that will be used before writing data batch
	// Otherwise all entries will be written consequentially
	maxBufferLen = batchSize * 1_000

	// Size of buffers of internal channels
	channelBuffer = 32

	filename1 = "test-file-1.txt"
	filename2 = "test-file-2.txt"
)

func BenchmarkFileBufferedForwarder(b *testing.B) {
	bfwd := NewBufferedForwarder(Config{
		BatchMaxLen:           batchSize,
		MaxForwarderBufferLen: maxBufferLen,
		Filename:              filename1,
		LastResortWriter:      nil, // In case of error writing file panic will occur
		ChannelsBuffer:        channelBuffer,
	})
	var wg sync.WaitGroup

	b.ResetTimer()

	wg.Add(goroutines)
	for j := 0; j < goroutines; j++ {
		go func() {
			bfwd.Write([]byte(strings.Repeat("HELLO WORLD FROM HERE LONG TEXT STARTS RIGHT HERE\n", stringsRepeat)))
			wg.Done()
		}()
	}
	wg.Wait()
	bfwd.Close()
}

func BenchmarkFileForwarder(b *testing.B) {
	fwd := newForwarder(
		filename2,
		nil, // In case of error writing file panic will occur
		maxBufferLen,
	)

	var wg sync.WaitGroup

	b.ResetTimer()

	wg.Add(goroutines)
	for j := 0; j < goroutines; j++ {
		go func() {
			fwd.Forward([]byte(strings.Repeat("HELLO WORLD FROM HERE LONG TEXT STARTS RIGHT HERE\n", stringsRepeat)))
			wg.Done()
		}()
	}
	wg.Wait()
}
