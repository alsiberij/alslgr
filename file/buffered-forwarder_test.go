package file

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"testing"
)

const (
	batchMaxLen           = 128
	forwarderMaxBufferLen = 10_000
	channelsBuffer        = 1

	maxNumbers = 10_000_000
	filename   = "test.txt"
)

func TestBufferedForwarder(t *testing.T) {
	bufferedFile := NewBufferedForwarder(Config{
		BatchMaxLen:           batchMaxLen,
		Filename:              filename,
		LastResortWriter:      os.Stdout,
		MaxForwarderBufferLen: forwarderMaxBufferLen,
		ChannelsBuffer:        channelsBuffer,
	})
	defer bufferedFile.Close()

	for i := 0; i < maxNumbers; i++ {
		bufferedFile.Write([]byte(fmt.Sprintf("%d\n", i)))
	}

	file, err := os.Open(filename)
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

	_ = os.Remove(filename)
}
