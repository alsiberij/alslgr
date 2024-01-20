package file

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"testing"
)

const (
	// testBatchSize is amount of data that will be aggregated into a batch
	testBatchSize = 100

	// testMaxBufferLen is maximum length of underlying buffer that will be used before writing data batch into a file
	// Otherwise all entries will be written consequentially
	testMaxBufferLen = 10000

	// channelsBuffer is a size of buffers of internal channels
	channelsBuffer = 32

	// testMaxNumbers is maximum amount of numbers to write in file
	testMaxNumbers = 10_000_000

	testFilename = "test.txt"
)

func TestBufferedForwarder(t *testing.T) {
	bufferedFile := NewBufferedForwarder(Config{
		BatchMaxLen:               testBatchSize,
		Filename:                  testFilename,
		LastResortWriter:          nil, // In case of error writing file panic will occur
		MaxBufferLenBeforeWriting: testMaxBufferLen,
		ChannelsBuffer:            channelsBuffer,
	})

	for i := 0; i < testMaxNumbers; i++ {
		bufferedFile.Write([]byte(fmt.Sprintf("%d\n", i)))
	}
	bufferedFile.Close()

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

	_ = file.Close()
	_ = os.Remove(testFilename)
}
