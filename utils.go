package alslgr

func broadcastToAllChannels[T any](mainCh <-chan T, chs ...chan T) {
	for range mainCh {
		for _, ch := range chs {
			ch <- *(new(T))
		}
	}
}
