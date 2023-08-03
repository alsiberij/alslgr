package alslgr

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	Concurrency       = 40000
	AutoDumpTestDelay = time.Millisecond * 50
)

type (
	TestDumper struct {
		Buf bytes.Buffer
	}
)

func (d *TestDumper) Dump(b []byte) {
	_, _ = d.Buf.Write(b)
}

type (
	Testcase struct {
		Name   string
		Cap    int
		Dumper Dumper
		Data   [][]byte
	}
)

func PrepareTests() []Testcase {
	return []Testcase{
		{
			Name:   "REGULAR WRITING",
			Cap:    1,
			Dumper: &TestDumper{},
			Data: [][]byte{
				[]byte("A"),
			},
		},
		{
			Name:   "WRITING ENTRY LARGER THAN BUFFER SIZE",
			Cap:    1,
			Dumper: &TestDumper{},
			Data: [][]byte{
				[]byte("AB"),
			},
		},
		{
			Name:   "WRITING ENTRY LARGER THAN BUFFER SIZE WITH FILLED BUFFER",
			Cap:    2,
			Dumper: &TestDumper{},
			Data: [][]byte{
				[]byte("A"), []byte("BCD"),
			},
		},
		{
			Name:   "WRITING ENTRIES WITH OVERFLOW",
			Cap:    2,
			Dumper: &TestDumper{},
			Data: [][]byte{
				[]byte("A"), []byte("B"), []byte("C"),
			},
		},
		{
			Name:   "WRITING ENTRIES WITH DOUBLE OVERFLOW",
			Cap:    2,
			Dumper: &TestDumper{},
			Data: [][]byte{
				[]byte("A"), []byte("B"), []byte("C"), []byte("D"), []byte("E"),
			},
		},
		{
			Name:   "DUMPER FORCED ERROR",
			Cap:    1,
			Dumper: &TestDumper{},
			Data: [][]byte{
				[]byte("A"), []byte("B"),
			},
		},
		{
			Name:   "DUMPER ERROR ON LARGE ENTRY",
			Cap:    4,
			Dumper: &TestDumper{},
			Data: [][]byte{
				[]byte("ABC"), []byte("DEFGH"),
			},
		},
	}
}

func mergeBytes(bs [][]byte) []byte {
	var length int

	for _, b := range bs {
		length += len(b)
	}

	result := make([]byte, 0, length)

	for _, b := range bs {
		result = append(result, b...)
	}

	return result
}

func TestWrite(t *testing.T) {
	tests := PrepareTests()

	for _, test := range tests {
		l := NewLogger(test.Cap, test.Dumper)
		l.Start()

		for _, data := range test.Data {
			l.Write(data)
		}

		l.ManualDump()

		givenResult := test.Dumper.(*TestDumper).Buf.Bytes()
		expectedResult := mergeBytes(test.Data)

		if !bytes.Equal(givenResult, expectedResult) {
			t.Errorf("TEST \"%s\" FAILED: EXPECTED DATA AFTER MANUAL DUMP \"%s\" GOT \"%s\"\n",
				test.Name, expectedResult, givenResult)
		}

		l.Stop()
	}
}

func TestAutoDump(t *testing.T) {
	d := &TestDumper{}

	l := NewLogger(10, d)
	l.Start()
	defer l.Stop()

	l.StartAutoDump(AutoDumpTestDelay)

	l.Write([]byte("AAA"))
	time.Sleep(AutoDumpTestDelay)

	l.Write([]byte("AAA"))
	time.Sleep(AutoDumpTestDelay)

	l.Write([]byte("AAA"))
	time.Sleep(AutoDumpTestDelay)

	l.Write([]byte("AAA"))
	time.Sleep(AutoDumpTestDelay * 2)

	givenResult := string((d.Buf).Bytes())

	if givenResult != "AAAAAAAAAAAA" {
		t.Errorf("TEST \"AUTO DUMP\" FAILED: EXPECTED DATA %s GOT %s\n", "AAAAAAAAAAAA", givenResult)
	}
}

func TestConcurrentWrite(t *testing.T) {
	d := &TestDumper{}

	l := NewLogger(100_000, d)
	l.Start()
	defer l.Stop()

	var wg sync.WaitGroup
	for i := 0; i < Concurrency; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, i int, l *Logger) {
			l.Write([]byte(fmt.Sprintf("%d| GOROUTINE WRITE\n", i)))
			wg.Done()
		}(&wg, i, &l)
	}

	wg.Wait()
	l.ManualDump()

	validate := regexp.MustCompile("^[0-9]+| GOROUTINE WRITE$").MatchString

	checkArr := [Concurrency]bool{}
	for {
		var s string
		s, err := d.Buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				t.Errorf("TEST \"CONCURRENT WRITE\" FAILED: EXPECTED READ ERROR \"nil\" GOT \"%v\"\n", err)
				return
			}
		}

		if !validate(s) {
			t.Errorf("TEST \"CONCURRENT WRITE\" FAILED: GOT INVALID ROW \"%s\"\n", s)
			continue
		}

		numStr, _, _ := strings.Cut(s, "|")

		var number int64
		number, err = strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			t.Errorf("TEST \"CONCURRENT WRITE\" FAILED: EXPECTED NUMBER CONVERTION ERROR \"nil\" GOT \"%v\"\n", err)
			continue
		}

		checkArr[number] = true
	}

	for i := range checkArr {
		if !checkArr[i] {
			t.Errorf("TEST \"CONCURRENT WRITE\" FAILED: LOST ROW \"%d\"\n", i)
		}
	}
}
