package alslgr

import (
	"context"
	"sync"
)

type (
	BufferedForwarder[B, T any] struct {
		workersWg           *sync.WaitGroup
		workersForwardersWg *sync.WaitGroup

		dataBatchProducer DataBatchProducer[B, T]
		dataForwarder     DataForwarder[B, T]

		dataCh      chan T
		dataBatchCh chan B

		doneCh                  chan struct{}
		reopenForwarderSignalCh <-chan struct{}
	}

	Config[B, T any] struct {
		DataBatchProducer        DataBatchProducer[B, T]
		DataForwarder            DataForwarder[B, T]
		ManualForwardingSignalCh <-chan struct{}
		ReopenForwarderCh        <-chan struct{}
		ChannelsBuffer           int
		BatchingConcurrency      int
		ForwardingConcurrency    int
	}
)

func NewBufferedForwarder[B, T any](config Config[B, T]) BufferedForwarder[B, T] {
	b := BufferedForwarder[B, T]{
		workersWg:               &sync.WaitGroup{},
		workersForwardersWg:     &sync.WaitGroup{},
		dataBatchProducer:       config.DataBatchProducer,
		dataForwarder:           config.DataForwarder,
		dataCh:                  make(chan T, config.ChannelsBuffer),
		dataBatchCh:             make(chan B, config.ChannelsBuffer),
		doneCh:                  make(chan struct{}),
		reopenForwarderSignalCh: config.ReopenForwarderCh,
	}

	manualForwardsChs := make([]chan struct{}, 0, config.BatchingConcurrency)

	for i := 0; i < config.BatchingConcurrency; i++ {
		b.workersWg.Add(1)

		manualForwardsCh := make(chan struct{}, 1)

		manualForwardsChs = append(manualForwardsChs, manualForwardsCh)
		go b.worker(manualForwardsCh)
	}

	mergeManualForwardChannels(config.ManualForwardingSignalCh, manualForwardsChs...)

	for i := 0; i < config.ForwardingConcurrency; i++ {
		b.workersForwardersWg.Add(1)
		go b.workerForwarder()
	}

	return b
}

func mergeManualForwardChannels(mainCh <-chan struct{}, chs ...chan struct{}) {
	for range mainCh {
		for _, ch := range chs {
			ch <- struct{}{}
		}
	}
}

func (l *BufferedForwarder[B, T]) worker(manualForwardCh <-chan struct{}) {
	defer l.workersWg.Done()

	batch := l.dataBatchProducer.NewDataBatch()

	for {
		select {
		case data, ok := <-l.dataCh:
			if !ok {
				l.dataBatchCh <- batch.Extract()
				return
			}
			batch.Append(data)
			if batch.ReadyToSend() {
				l.dataBatchCh <- batch.Extract()
				batch.Reset()
			}
		case <-l.doneCh:
			l.dataBatchCh <- batch.Extract()
			return
		case <-manualForwardCh:
			l.dataBatchCh <- batch.Extract()
			batch.Reset()
		}
	}
}

func (l *BufferedForwarder[B, T]) workerForwarder() {
	defer l.workersForwardersWg.Done()
	defer l.handleRemainingData()

	for {
		select {
		case batch, ok := <-l.dataBatchCh:
			if !ok {
				return
			}
			l.dataForwarder.ForwardDataBatch(batch)
		case <-l.reopenForwarderSignalCh:
			l.dataForwarder.Reopen()
		}
	}
}

func (l *BufferedForwarder[B, T]) handleRemainingData() {
	for data := range l.dataCh {
		l.dataForwarder.ForwardData(data)
	}
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
	l.workersWg.Wait()

	close(l.dataBatchCh)
	l.workersForwardersWg.Wait()

	l.dataForwarder.Close()
}
