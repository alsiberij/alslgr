package alslgr

import (
	"context"
	"time"
)

func repeatOpWorker(ctx context.Context, interval time.Duration, errCh chan<- error, op func() error) {
	defer close(errCh)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			select {
			case errCh <- op():
			default:
			}
		}
	}
}
