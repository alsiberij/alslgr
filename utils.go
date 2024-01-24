package alslgr

import (
	"os"
	"syscall"
	"time"
)

// AutoWritingBatches starts ticking goroutine that will send signals to send accumulated batches. Consider using
// rare ticking in case of slow writing. Function returns channel that can be used for cancelling signalling
// goroutine by closing it. After such cancelling internal channels are not affected, so you can start a new ticker
func AutoWritingBatches[B, T any](bw *BatchedWriter[B, T], t time.Duration) chan<- struct{} {
	doneCh := make(chan struct{}, 1)
	go tick(t, bw.saveBatchesCh, doneCh)
	return doneCh
}

func tick(interval time.Duration, sigCh chan<- struct{}, doneCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			select {
			case sigCh <- struct{}{}:
			default:
			}
		case _, ok := <-doneCh:
			if !ok {
				return
			}
		}
	}
}

// ResetWriterOnSigHup is reading from sigHupCh waiting for the syscall.SIGHUP to reset the state
// of internal Writer of bw. Function returns channel that can be used for cancelling signalling
// goroutine by closing it. After such cancelling internal channels are not affected, so you call this function again
func ResetWriterOnSigHup[B, T any](bw *BatchedWriter[B, T], sigHupCh <-chan os.Signal) chan<- struct{} {
	doneCh := make(chan struct{}, 1)
	go sigHupHandler(sigHupCh, bw.resetWriterCh, doneCh)
	return doneCh
}

func sigHupHandler(sigHupCh <-chan os.Signal, sigCh chan<- struct{}, doneCh <-chan struct{}) {
	for {
		select {
		case sig, ok := <-sigHupCh:
			if !ok {
				return
			}

			switch sig {
			case syscall.SIGHUP:
				select {
				case sigCh <- struct{}{}:
				default:
				}
			}
		case _, ok := <-doneCh:
			if !ok {
				return
			}
		}
	}
}
