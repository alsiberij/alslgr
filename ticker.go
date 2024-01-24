package alslgr

import "time"

// SaveBatchesByTicker starts ticking goroutine that will send signals to send accumulated batches. Consider using
// rare ticking in case of slow writing. Function returns channel that can be used for cancelling signalling
// goroutine by closing it. After such cancelling internal channels are not affected so you can start a new ticker
func SaveBatchesByTicker[B, T any](bw *BatchedWriter[B, T], t time.Duration) chan<- struct{} {
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
			sigCh <- struct{}{}
		case _, ok := <-doneCh:
			if !ok {
				return
			}
		}
	}
}
