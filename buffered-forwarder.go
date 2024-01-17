package alslgr

import (
	"sync"
)

type (
	BufferedForwarder[B, T any] struct {
		workersWg           *sync.WaitGroup
		workersForwardersWg *sync.WaitGroup

		dataBatchProducer DataBatchProducer[B, T]
		dataForwarder     DataForwarder[B, T]

		dataCh      chan T
		dataBatchCh chan DataBatch[B, T]

		doneCh                   chan struct{}
		manualForwardingSignalCh <-chan struct{}
		reopenForwarderSignalCh  <-chan struct{}
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
		workersWg:                &sync.WaitGroup{},
		workersForwardersWg:      &sync.WaitGroup{},
		dataBatchProducer:        config.DataBatchProducer,
		dataForwarder:            config.DataForwarder,
		dataCh:                   make(chan T, config.ChannelsBuffer),
		dataBatchCh:              make(chan DataBatch[B, T], config.ChannelsBuffer),
		doneCh:                   make(chan struct{}),
		manualForwardingSignalCh: config.ManualForwardingSignalCh,
		reopenForwarderSignalCh:  config.ReopenForwarderCh,
	}

	for i := 0; i < config.BatchingConcurrency; i++ {
		b.workersWg.Add(1)
		go b.worker()
	}

	for i := 0; i < config.ForwardingConcurrency; i++ {
		b.workersForwardersWg.Add(1)
		go b.workerForwarder()
	}

	return b
}

func (l *BufferedForwarder[B, T]) worker() {
	defer l.workersWg.Done()

	batch := l.dataBatchProducer.NewDataBatch()

	for {
		select {
		case data, ok := <-l.dataCh:
			if !ok {
				l.dataBatchCh <- batch
				return
			}
			batch.Append(data)
			if batch.ReadyToSend() {
				l.dataBatchCh <- batch
				batch = l.dataBatchProducer.NewDataBatch()
			}
		case <-l.doneCh:
			l.dataBatchCh <- batch
			return
		case <-l.manualForwardingSignalCh:
			l.dataBatchCh <- batch
			batch = l.dataBatchProducer.NewDataBatch()
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
			l.dataForwarder.ForwardDataBatch(batch.Extract())
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

func (l *BufferedForwarder[B, T]) Close() {
	close(l.doneCh)
	close(l.dataCh)
	l.workersWg.Wait()

	close(l.dataBatchCh)
	l.workersForwardersWg.Wait()

	l.dataForwarder.Close()
}
