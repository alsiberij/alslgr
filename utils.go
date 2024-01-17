package alslgr

import "time"

func TimedForwarding(interval time.Duration, doneCh <-chan struct{}) <-chan struct{} {
	sigCh := make(chan struct{})
	go tick(interval, sigCh, doneCh)
	return sigCh
}

func tick(interval time.Duration, sigCh chan<- struct{}, doneCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer close(sigCh)

	for {
		select {
		case <-ticker.C:
			sigCh <- struct{}{}
		case <-doneCh:
			return
		}
	}
}
