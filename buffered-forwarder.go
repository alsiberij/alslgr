package alslgr

import (
	"context"
	"sync"
)

type (
	BufferedForwarder[B, T any] struct {
		workersBatchingWg   *sync.WaitGroup
		workersForwardingWg *sync.WaitGroup

		batchProducer BatchProducer[B, T]
		forwarder     Forwarder[B, T]

		dataCh  chan T
		batchCh chan B

		doneCh           chan struct{}
		resetForwarderCh <-chan struct{}
	}

	Config[B, T any] struct {
		BatchProducer         BatchProducer[B, T]
		Forwarder             Forwarder[B, T]
		ManualForwardingCh    <-chan struct{}
		ResetForwarderCh      <-chan struct{}
		ChannelsBuffer        int
		BatchingConcurrency   int
		ForwardingConcurrency int
	}
)

func NewBufferedForwarder[B, T any](config Config[B, T]) BufferedForwarder[B, T] {
	b := BufferedForwarder[B, T]{
		workersBatchingWg:   &sync.WaitGroup{},
		workersForwardingWg: &sync.WaitGroup{},
		batchProducer:       config.BatchProducer,
		forwarder:           config.Forwarder,
		dataCh:              make(chan T, config.ChannelsBuffer),
		batchCh:             make(chan B, config.ChannelsBuffer),
		doneCh:              make(chan struct{}),
		resetForwarderCh:    config.ResetForwarderCh,
	}

	manualForwardsChs := make([]chan struct{}, 0, config.BatchingConcurrency)

	b.workersBatchingWg.Add(config.BatchingConcurrency)
	for i := 0; i < config.BatchingConcurrency; i++ {
		manualForwardsCh := make(chan struct{}, 1)
		manualForwardsChs = append(manualForwardsChs, manualForwardsCh)
		go b.workerBatching(manualForwardsCh)
	}

	go mergeManualForwardChannels(config.ManualForwardingCh, manualForwardsChs...)

	b.workersForwardingWg.Add(config.ForwardingConcurrency)
	for i := 0; i < config.ForwardingConcurrency; i++ {
		go b.workerForwarding()
	}

	return b
}

func (l *BufferedForwarder[B, T]) Write(data T) {
	l.dataCh <- data
}

func (l *BufferedForwarder[B, T]) WriteCtx(ctx context.Context, data T) error {
	select {
	case l.dataCh <- data:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *BufferedForwarder[B, T]) Close() {
	close(l.doneCh)
	close(l.dataCh)
	l.workersBatchingWg.Wait()

	close(l.batchCh)
	l.workersForwardingWg.Wait()

	l.forwarder.Close()
}

func (l *BufferedForwarder[B, T]) workerBatching(manualForwardCh <-chan struct{}) {
	defer l.workersBatchingWg.Done()

	batch := l.batchProducer.NewBatch()

	for {
		select {
		case data, ok := <-l.dataCh:
			if !ok {
				l.batchCh <- batch.Extract()
				return
			}
			batch.Append(data)
			if batch.IsFull() {
				l.batchCh <- batch.Extract()
			}
		case <-l.doneCh:
			l.batchCh <- batch.Extract()
			return
		case <-manualForwardCh:
			l.batchCh <- batch.Extract()
		}
	}
}

func (l *BufferedForwarder[B, T]) workerForwarding() {
	defer l.workersForwardingWg.Done()
	defer l.handleRemainingData()

	for {
		select {
		case batch, ok := <-l.batchCh:
			if !ok {
				return
			}
			l.forwarder.ForwardBatch(batch)
		case <-l.resetForwarderCh:
			l.forwarder.Reset()
		}
	}
}

func (l *BufferedForwarder[B, T]) handleRemainingData() {
	for data := range l.dataCh {
		l.forwarder.Forward(data)
	}
}
