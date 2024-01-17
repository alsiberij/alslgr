package main

import (
	"github.com/alsiberij/alslgr/v3/files"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP)

	doneCh := make(chan struct{})
	defer close(doneCh)

	fileForwarder := files.NewBufferedForwarder(files.Config{
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
