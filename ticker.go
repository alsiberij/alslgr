package alslgr

import "time"

// NewTicker returns read-only channel in which empty struct will be written after each tick. Not that doneCh can be
// used to stop sending goroutine
func NewTicker(t time.Duration, doneCh <-chan struct{}) <-chan struct{} {
	sigCh := make(chan struct{})
	go tick(t, sigCh, doneCh)
	return sigCh
}
