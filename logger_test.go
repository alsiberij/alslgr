package alslgr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type (
	TestDumper struct {
		Buf        bytes.Buffer
		NoErr      bool
		ErrCounter int
		Errs       []error
	}
)

const (
	Concurrency       = 1000
	AutoDumpTestDelay = time.Millisecond * 300
)

var (
	forcedError            = errors.New("forced error")
	errInvalidErrorCounter = errors.New("invalid error counter")
)

func (d *TestDumper) Dump(b []byte) error {
	if !d.NoErr {
		if d.ErrCounter >= len(d.Errs) {
			return errInvalidErrorCounter
		}

		err := d.Errs[d.ErrCounter]
		d.ErrCounter++

		if err != nil {
			return err
		}
	}
	_, _ = d.Buf.Write(b)
	return nil
}

type (
	Testcase struct {
		Name                       string
		Cap                        int
		Dumper                     Dumper
		Data                       [][]byte
		DumpedDataBeforeManualDump []byte
		ExpectedWriteErrs          []error
	}
)

func PrepareTests() []Testcase {
	return []Testcase{
		{
			Name: "REGULAR WRITING",
			Cap:  1,
			Dumper: &TestDumper{
				Errs: []error{nil},
			},
			Data: [][]byte{
				[]byte("A"),
			},
			DumpedDataBeforeManualDump: []byte{},
			ExpectedWriteErrs:          []error{nil},
		},
		{
			Name: "WRITING ENTRY LARGER THAN BUFFER SIZE",
			Cap:  1,
			Dumper: &TestDumper{
				Errs: []error{nil, nil, nil},
			},
			Data: [][]byte{
				[]byte("AB"),
			},
			DumpedDataBeforeManualDump: []byte("AB"),
			ExpectedWriteErrs:          []error{nil},
		},
		{
			Name: "WRITING ENTRY LARGER THAN BUFFER SIZE WITH FILLED BUFFER",
			Cap:  2,
			Dumper: &TestDumper{
				Errs: []error{nil, nil, nil},
			},
			Data: [][]byte{
				[]byte("A"), []byte("BCD"),
			},
			DumpedDataBeforeManualDump: []byte("ABCD"),
			ExpectedWriteErrs:          []error{nil, nil},
		},
		{
			Name: "WRITING ENTRIES WITH OVERFLOW",
			Cap:  2,
			Dumper: &TestDumper{
				Errs: []error{nil, nil},
			},
			Data: [][]byte{
				[]byte("A"), []byte("B"), []byte("C"),
			},
			DumpedDataBeforeManualDump: []byte("AB"),
			ExpectedWriteErrs:          []error{nil, nil, nil},
		},
		{
			Name: "WRITING ENTRIES WITH DOUBLE OVERFLOW",
			Cap:  2,
			Dumper: &TestDumper{
				Errs: []error{nil, nil, nil},
			},
			Data: [][]byte{
				[]byte("A"), []byte("B"), []byte("C"), []byte("D"), []byte("E"),
			},
			DumpedDataBeforeManualDump: []byte("ABCD"),
			ExpectedWriteErrs:          []error{nil, nil, nil, nil, nil},
		},
		{
			Name: "DUMPER FORCED ERROR",
			Cap:  1,
			Dumper: &TestDumper{
				Errs: []error{forcedError, forcedError},
			},
			Data: [][]byte{
				[]byte("A"), []byte("B"),
			},
			DumpedDataBeforeManualDump: []byte{},
			ExpectedWriteErrs:          []error{nil, forcedError},
		},
		{
			Name: "DUMPER ERROR ON LARGE ENTRY",
			Cap:  4,
			Dumper: &TestDumper{
				Errs: []error{nil, forcedError, forcedError},
			},
			Data: [][]byte{
				[]byte("ABC"), []byte("DEFGH"),
			},
			DumpedDataBeforeManualDump: []byte("ABC"),
			ExpectedWriteErrs:          []error{nil, forcedError},
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

func TestLogger(t *testing.T) {
	tests := PrepareTests()

	for _, test := range tests {
		l := NewLogger(test.Cap, test.Dumper)

		for i, data := range test.Data {
			_, err := l.Write(data)
			if !errors.Is(err, test.ExpectedWriteErrs[i]) {
				t.Errorf("TEST \"%s\" FAILED: EXPECTED WRITE ERROR \"%v\" GOT \"%v\"\n", test.Name, test.ExpectedWriteErrs[i], err)
			}
		}

		dataBeforeDump := test.Dumper.(*TestDumper).Buf.Bytes()

		if !bytes.Equal(dataBeforeDump, test.DumpedDataBeforeManualDump) {
			t.Errorf("TEST \"%s\" FAILED: EXPECTED DATA BEFORE MANUAL DUMP \"%s\" GOT \"%s\"\n",
				test.Name, test.DumpedDataBeforeManualDump, dataBeforeDump)
		}

		err := l.DumpBuffer()
		expectedErr := test.Dumper.(*TestDumper).Errs[len(test.Dumper.(*TestDumper).Errs)-1]
		if !errors.Is(err, expectedErr) {
			t.Errorf("TEST \"%s\" FAILED: EXPECTED DUMP ERROR \"%v\" GOT %v\n", test.Name, expectedErr, err)
		}
		if err != nil {
			continue
		}

		givenResult := test.Dumper.(*TestDumper).Buf.Bytes()
		expectedResult := mergeBytes(test.Data)
		if !bytes.Equal(givenResult, expectedResult) {
			t.Errorf("TEST \"%s\" FAILED: EXPECTED DATA AFTER MANUAL DUMP \"%s\" GOT \"%s\"\n",
				test.Name, expectedResult, givenResult)
		}
	}
}

func TestAutoDump(t *testing.T) {
	d := &TestDumper{
		NoErr: true,
	}
	l := NewLogger(1<<3, d)

	_, err := l.Write([]byte("A"))
	if err != nil {
		t.Errorf("TEST \"AUTO DUMP\" FAILED: EXPECTED WRITE ERROR \"nil\" GOT \"%v\"\n", err)
	}

	errCh, cancel := l.AutoDumpBuffer(AutoDumpTestDelay)

	go func() {
		for dumpErr := range errCh {
			if dumpErr != nil {
				t.Errorf("TEST \"AUTO DUMP\" FAILED: EXPECTED ASYNC DUMP ERROR \"nil\" GOT \"%v\"\n", dumpErr)
			}
		}
	}()

	_, err = l.Write([]byte("A"))
	if err != nil {
		t.Errorf("TEST \"AUTO DUMP\" FAILED: EXPECTED WRITE ERROR \"nil\" GOT \"%v\"\n", err)
	}
	time.Sleep(AutoDumpTestDelay)

	_, err = l.Write([]byte("A"))
	if err != nil {
		t.Errorf("TEST \"AUTO DUMP\" FAILED: EXPECTED WRITE ERROR \"nil\" GOT \"%v\"\n", err)
	}
	time.Sleep(AutoDumpTestDelay * 2)

	cancel()

	givenResult := string((d.Buf).Bytes())
	if givenResult != "AAA" {
		t.Errorf("TEST \"AUTO DUMP\" FAILED: EXPECTED DATA %s GOT %s\n", "AAA", givenResult)
	}
}

func TestConcurrentWrite(t *testing.T) {
	d := &TestDumper{
		NoErr: true,
	}
	l := NewLogger(1<<12, d)

	var wg sync.WaitGroup
	for i := 0; i < Concurrency; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, i int, l Logger, t *testing.T) {
			defer wg.Done()

			_, err := l.Write([]byte(fmt.Sprintf("%d| GOROUTINE WRITE\n", i)))
			if err != nil {
				t.Errorf("TEST \"CONCURRENT WRITE\" FAILED: EXPECTED WRITE ERROR \"nil\" GOT \"%v\"\n", err)
			}
		}(&wg, i, l, t)
	}

	wg.Wait()

	err := l.DumpBuffer()
	if err != nil {
		t.Errorf("TEST \"CONCURRENT WRITE\" FAILED: EXPECTED DUMP ERROR \"nil\" GOT \"%v\"\n", err)
		return
	}

	validate := regexp.MustCompile("^[0-9]+| GOROUTINE WRITE$").MatchString

	checkArr := [Concurrency]bool{}
	for {
		var s string
		s, err = d.Buf.ReadString('\n')
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
