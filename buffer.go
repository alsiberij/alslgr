package alslgr

import (
	"sync"
)

type (
	Buffer[B, T any] struct {
		wg *sync.WaitGroup

		batchProducer BatchProducer[B, T]
		dataForwarder DataForwarder[B, T]

		dataCh  chan T
		batchCh chan Batch[B, T]

		doneCh            chan struct{}
		manualForwardCh   chan struct{}
		reopenForwarderCh chan struct{}
	}
)

func NewBuffer[B, T any](batchProducer BatchProducer[B, T], dataForwarder DataForwarder[B, T],
	manualForwardSignalCh, reopenForwarderCh chan struct{}, channelsBuffer,
	batchingConcurrency, forwardingConcurrency int) Buffer[B, T] {

	b := Buffer[B, T]{
		wg:                &sync.WaitGroup{},
		batchProducer:     batchProducer,
		dataForwarder:     dataForwarder,
		dataCh:            make(chan T, channelsBuffer),
		batchCh:           make(chan Batch[B, T], channelsBuffer),
		doneCh:            make(chan struct{}),
		manualForwardCh:   manualForwardSignalCh,
		reopenForwarderCh: reopenForwarderCh,
	}

	for i := 0; i < forwardingConcurrency; i++ {
		b.wg.Add(1)
		go b.worker()
	}

	for i := 0; i < batchingConcurrency; i++ {
		b.wg.Add(1)
		go b.workerForwarder()
	}

	return b
}

func (l *Buffer[B, T]) worker() {
	defer l.wg.Done()
	defer close(l.batchCh)

	batch := l.batchProducer.NewBatch()

	for {
		select {
		case data, ok := <-l.dataCh:
			if !ok {
				return
			}
			batch.Append(data)
			if batch.ReadyToSend() {
				l.batchCh <- batch
				batch = l.batchProducer.NewBatch()
			}
		case <-l.doneCh:
			l.batchCh <- batch
			return
		case <-l.manualForwardCh:
			l.batchCh <- batch
			batch = l.batchProducer.NewBatch()
		}
	}
}

func (l *Buffer[B, T]) workerForwarder() {
	defer l.wg.Done()
	defer l.handleRemainingData()

	for {
		select {
		case batch, ok := <-l.batchCh:
			if !ok {
				return
			}
			l.dataForwarder.ForwardBatch(batch.Extract())
		case <-l.reopenForwarderCh:
			l.dataForwarder.Reopen()
		}
	}
}

func (l *Buffer[B, T]) handleRemainingData() {
	for data := range l.dataCh {
		l.dataForwarder.Forward(data)
	}
}

func (l *Buffer[B, T]) Write(data T) {
	l.dataCh <- data
}

func (l *Buffer[B, T]) Close() {
	close(l.doneCh)
	close(l.dataCh)
	l.wg.Wait()
	l.dataForwarder.Close()
}
