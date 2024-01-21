package alslgr

import (
	"context"
	"sync"
)

type (
	// BatchedWriter is some kind of generic buffer that accumulates data batches before writing it in destination
	// in cases where regular writing is expensive. BatchedWriter also supports a signal channel for immediate
	// writing all batches as well as graceful shutdown with Close method.
	// Struct is typed with B and T, where T is data itself and B is a batch of T.
	// Consider passing a copy of data in Write or WriteCtx if data can be modified outside
	// It is now allowed to call any methods after calling Close.
	BatchedWriter[B, T any] struct {
		// workersBatchingWg is used for waiting until all workerBatch are stopped
		workersBatchingWg *sync.WaitGroup

		// workersWritingWg is used for waiting until all workerWrite are stopped.
		// It is separated from workersBatchingWg because all the workerBatch must be stopped before
		// stopping any workerWrite. Otherwise, panic may occur (sending to a close channel)
		workersWritingWg *sync.WaitGroup

		// batchProducer is used to initialize Batch inside every workerBatch
		batchProducer BatchProducer[B, T]

		// writer is used by workerWrite for writing data to destination
		writer Writer[B, T]

		// dataCh is used by workerBatch to accumulate batches of data. Methods Write and WriteCtx write in it
		// and workerBatch are read from it
		dataCh chan T

		// batchCh is used for communication between workerBatch and workerWrite. After batch is ready to send,
		// workerBatch calls Batch.Extract and sends returned value to this channel. After that, workerWrite reads
		// sent data from it and calls Writer.WriteBatch
		batchCh chan B

		// doneCh is used for graceful shutdown of BatchedWriter. It is closing when Close is called
		doneCh chan struct{}

		// resetWriterCh is used by workerWrite for safe calling of Writer.Reset when needed
		resetWriterCh <-chan struct{}
	}

	// Config is a set of required parameters for initializing BatchedWriter
	Config[B, T any] struct {
		BatchProducer   BatchProducer[B, T]
		Writer          Writer[B, T]
		ManualWritingCh <-chan struct{}
		ResetWriterCh   <-chan struct{}
		ChannelsBuffer  int
		BatchingWorkers int
		WritingWorkers  int
	}
)

func NewBatchedWriter[B, T any](config Config[B, T]) BatchedWriter[B, T] {
	b := BatchedWriter[B, T]{
		workersBatchingWg: &sync.WaitGroup{},
		workersWritingWg:  &sync.WaitGroup{},
		batchProducer:     config.BatchProducer,
		writer:            config.Writer,
		dataCh:            make(chan T, config.ChannelsBuffer),
		batchCh:           make(chan B, config.ChannelsBuffer),
		doneCh:            make(chan struct{}),
		resetWriterCh:     config.ResetWriterCh,
	}

	manualForwardsChs := make([]chan struct{}, 0, config.BatchingWorkers)

	b.workersBatchingWg.Add(config.BatchingWorkers)
	for i := 0; i < config.BatchingWorkers; i++ {
		manualForwardsCh := make(chan struct{}, 1)
		manualForwardsChs = append(manualForwardsChs, manualForwardsCh)
		go b.workerBatch(manualForwardsCh)
	}

	go mergeChannels[struct{}](config.ManualWritingCh, manualForwardsChs...)

	b.workersWritingWg.Add(config.WritingWorkers)
	for i := 0; i < config.WritingWorkers; i++ {
		go b.workerWrite()
	}

	return b
}

func (l *BatchedWriter[B, T]) Write(data T) {
	l.dataCh <- data
}

func (l *BatchedWriter[B, T]) WriteCtx(ctx context.Context, data T) error {
	select {
	case l.dataCh <- data:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *BatchedWriter[B, T]) Close() {
	close(l.doneCh)
	close(l.dataCh)
	l.workersBatchingWg.Wait()

	close(l.batchCh)
	l.workersWritingWg.Wait()

	l.writer.Close()
}

func (l *BatchedWriter[B, T]) workerBatch(manualForwardCh <-chan struct{}) {
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

func (l *BatchedWriter[B, T]) workerWrite() {
	defer l.workersWritingWg.Done()
	defer l.handleRemainingData()

	for {
		select {
		case batch, ok := <-l.batchCh:
			if !ok {
				return
			}
			l.writer.WriteBatch(batch)
		case <-l.resetWriterCh:
			l.writer.Reset()
		}
	}
}

func (l *BatchedWriter[B, T]) handleRemainingData() {
	for data := range l.dataCh {
		l.writer.Write(data)
	}
}
