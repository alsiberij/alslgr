package file

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"testing"
)

const (
	// Amount of data that will be aggregated into a batch
	testBatchSize = 100

	// Maximum length of underlying buffer that will be used before writing data batch into a file
	// Otherwise all entries will be written consequentially
	testMaxBufferLen = 10_000

	// Size of buffers of internal channels
	channelsBuffer = 1

	// Maximum numbers to write in file
	testMaxNumbers = 10_000_000

	testFilename = "test.txt"
)

func TestBufferedForwarder(t *testing.T) {
	bufferedFile := NewBufferedForwarder(Config{
		BatchMaxLen:           testBatchSize,
		Filename:              testFilename,
		LastResortWriter:      os.Stdout,
		MaxForwarderBufferLen: testMaxBufferLen,
		ChannelsBuffer:        channelsBuffer,
	})
	defer bufferedFile.Close()

	for i := 0; i < testMaxNumbers; i++ {
		bufferedFile.Write([]byte(fmt.Sprintf("%d\n", i)))
	}

	file, err := os.Open(testFilename)
	if err != nil {
		t.Fatal(err)
	}

	s := bufio.NewScanner(file)

	for i := int64(0); s.Scan(); i++ {
		n, _ := strconv.ParseInt(s.Text(), 10, 64)
		if n != i {
			t.Fatalf("Expected %d, got %d", i, n)
		}
	}

	_ = os.Remove(testFilename)
}
