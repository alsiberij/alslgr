package main

import (
	"fmt"
	"github.com/alsiberij/alslgr/v3/file"
	"github.com/alsiberij/alslgr/v3/stdout"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func runFileForwarder() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP)

	doneCh := make(chan struct{})
	defer close(doneCh)

	fileForwarder := file.NewBufferedForwarder(file.Config{
		BatchMaxLen:           16,
		Filename:              "log.txt",
		LastResortWriter:      os.Stdout,
		MaxForwarderBufferLen: 100,
		TimedForwardInterval:  time.Second,
		SighupCh:              sigCh,
		TimedForwardingDoneCh: doneCh,
		ReopenForwarderDoneCh: doneCh,
		ChannelsBuffer:        1,
	})
	defer fileForwarder.Close()

	var wg sync.WaitGroup
	for i := 0; i < 100_000; i++ {
		wg.Add(1)
		go func() {
			fileForwarder.Write([]byte("HELLO FROM HERE!!!\n"))
			wg.Done()
		}()
	}
	wg.Wait()
}

func runStdoutForwarder() {
	stdoutForwarder := stdout.NewBufferedForwarder(stdout.Config{
		BatchMaxLen:           32,
		ChannelsBuffer:        0,
		BatchingConcurrency:   16,
		ForwardingConcurrency: 8,
	})
	defer stdoutForwarder.Close()

	var wg sync.WaitGroup
	for i := 0; i < 100_000; i++ {
		wg.Add(1)
		go func(i int) {
			stdoutForwarder.Write([]byte(fmt.Sprintf("%d\n", i)))
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func main() {
	runFileForwarder()
	runStdoutForwarder()
}
