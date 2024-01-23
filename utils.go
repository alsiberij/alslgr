package alslgr

import "time"

func broadcastToAll[T any](mainCh <-chan T, chs ...chan T) {
	for range mainCh {
		for _, ch := range chs {
			ch <- *(new(T))
		}
	}
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
