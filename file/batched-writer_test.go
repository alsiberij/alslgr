package file

import (
	"bufio"
	"context"
	"fmt"
	"github.com/alsiberij/alslgr/v3"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
)

const (
	// goroutinesWriting is a number of goroutines that will write data in BatchedWriter
	goroutinesWriting = 100

	// dataRepeatPerGoroutineWriting is a number of repeats of data inside every writing goroutine
	dataRepeatPerGoroutineWriting = 10000

	// batchMaxLength is maximum length of internal []byte buffer
	batchMaxLength = 1_000_000

	fileName = "batched-writer_test.txt"
)

func TestBatchedWriter(t *testing.T) {
	sigHupCh := make(chan os.Signal)

	bw, err := NewBatchedWriter(batchMaxLength, fileName, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}
	// To test reopening file
	defer close(alslgr.ResetWriterOnSigHup(&bw, sigHupCh))

	var wg sync.WaitGroup
	for i := 0; i < goroutinesWriting; i++ {
		wg.Add(1)
		go func(i int) {
			for j := 0; j < dataRepeatPerGoroutineWriting; j++ {
				// Writing some string and number of goroutine with repeat index
				// This data will be used to check data integrity
				bw.Write([]byte(fmt.Sprintf("HELLO FROM #%d|%d\n", i, j)))
			}
			wg.Done()
		}(i)
	}

	// Reopening file when possible
	go func() {
		sigHupCh <- syscall.SIGHUP
	}()

	// To test Pack flow inside BatchedWriter
	_ = bw.WriteCtx(context.Background(), []byte(strings.Repeat("_", batchMaxLength)+"#_|_\n"))

	wg.Wait()

	bw.Close()

	f, err := os.Open(fileName)
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 0, batchMaxLength*2), batchMaxLength*2)

	// True in checkList[i][j] means that i'th goroutine successfully wrote j'th repeat of data
	checkList := make([][]bool, goroutinesWriting)
	for g := 0; g < goroutinesWriting; g++ {
		checkList[g] = make([]bool, dataRepeatPerGoroutineWriting)
	}

	for s.Scan() {
		numbers := strings.Split(strings.Split(s.Text(), "#")[1], "|")
		rawG, rawR := numbers[0], numbers[1]

		// Skipping Pack flow check
		if rawG == "_" && rawR == "_" {
			continue
		}

		g, _ := strconv.Atoi(rawG)
		r, _ := strconv.Atoi(rawR)

		checkList[g][r] = true
	}

	for i, l := range checkList {
		for j, ok := range l {
			if !ok {
				t.Fatalf("Expected true for (%dg, %dr) got false", i, j)
			}
		}
	}

	_ = f.Close()
	_ = os.Remove(fileName)
}
