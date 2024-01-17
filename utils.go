package alslgr

import "time"

func TimedForwarding(interval time.Duration, doneCh <-chan struct{}) chan<- struct{} {
	tickCh := make(chan<- struct{})
	go tick(interval, tickCh, doneCh)
	return tickCh
}

func tick(interval time.Duration, sigCh chan<- struct{}, doneCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sigCh <- struct{}{}
		case <-doneCh:
			return
		}
	}
}
