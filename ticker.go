package alslgr

import "time"

// NewTicker returns a read-only channel in which empty struct will be written after each tick.
// Note that doneCh can be used to stop sending goroutine
func NewTicker(t time.Duration, doneCh <-chan struct{}) <-chan struct{} {
	sigCh := make(chan struct{})
	go tick(t, sigCh, doneCh)
	return sigCh
}
