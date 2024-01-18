package alslgr

func mergeManualForwardChannels(mainCh <-chan struct{}, chs ...chan struct{}) {
	for range mainCh {
		for _, ch := range chs {
			ch <- struct{}{}
		}
	}
}
