package alslgr

import (
	"context"
	"sync"
)

type (
	// BatchedWriter is some kind of generic buffer that accumulates data batches before writing it in destination
	// in cases where regular writing might be expensive. BatchedWriter also supports a signal channel for immediate
	// writing all batches as well as graceful shutdown with Close method.
	// Struct is typed with B and T, where T is data to write, and B is a batch of T's.
	// Consider passing a copy of data in Write or WriteCtx if data can be modified outside.
	// It is not allowed to call any methods after calling Close.
	BatchedWriter[B, T any] struct {
		// workersBatchingWg is used for waiting until all workerBatch are stopped
		workersBatchingWg *sync.WaitGroup

		// workersWritingWg is used for waiting until all workerWrite are stopped.
		// It is separated from workersBatchingWg because all the workerBatch must be stopped before
		// stopping any workerWrite. Otherwise, panic may occur (sending to a closed channel)
		workersWritingWg *sync.WaitGroup

		// batchProducer is used to initialize Batch inside every workerBatch
		batchProducer BatchProducer[B, T]

		// writer is used by workerWrite for writing data
		writer Writer[B, T]

		// dataCh is used by workerBatch to accumulate batches of data. Methods Write and WriteCtx write in it
		// and workerBatch read from it
		dataCh chan T

		// batchCh is used for communication between workerBatch and workerWrite. After batch is ready to send,
		// workerBatch calls Batch.Extract and sends returned value to this channel. After that, workerWrite reads
		// sent data and calls Writer.WriteBatch
		batchCh chan B

		// doneCh is used for graceful shutdown of BatchedWriter as well as free its own resources. It is closing
		// when Close is called
		doneCh chan struct{}

		// saveBatchesCh is used to manually write accumulated batches whether they are full or not.
		// Internally, the signal received from this channel will be split and sent to every worker indicating
		// that there is time to send data. Method SaveBatches will send to this channel, so don't close it if
		// you consider using SaveBatches
		saveBatchesCh chan struct{}

		// resetWriterCh is used by workerWrite for safe calling of Writer.Reset when needed
		resetWriterCh <-chan struct{}
	}

	// BatchedWriterConfig is a set of required parameters for initializing BatchedWriter
	BatchedWriterConfig[B, T any] struct {
		// BatchProducer is needed to make new Batch inside every worker.
		// Note that the method of BatchProducer can be called from different goroutines if Workers > 0
		BatchProducer BatchProducer[B, T]

		// Writer is needed to write data in destination.
		// Not that any method of Writer must be safe to call from different goroutines if Workers > 0
		Writer Writer[B, T]

		// SaveBatchesCh is needed to manually write all accumulated batches. Internally, the signal received
		// from this channel will be split and sent to every worker indicating that there is time to send data
		SaveBatchesCh chan struct{}

		// ResetWriterCh is needed to signal that Writer.Reset must be called
		ResetWriterCh <-chan struct{}

		// ChannelBuffer is needed to initialize buffer of underlying channels
		ChannelBuffer int

		// Workers is the number of workers that will be spawned for accumulating and writing data.
		// Note that having more than one worker will be more performant only in cases where exclusive access to
		// Writer is not needed, and keep in mind that methods of Writer could be called from different goroutines.
		// Otherwise, it is recommended to have only one worker. Also, having one worker is necessary when
		// consequential writing is needed, for example, writing in file
		Workers int
	}
)

func NewBatchedWriter[B, T any](config BatchedWriterConfig[B, T]) BatchedWriter[B, T] {
	b := BatchedWriter[B, T]{
		workersBatchingWg: &sync.WaitGroup{},
		workersWritingWg:  &sync.WaitGroup{},
		batchProducer:     config.BatchProducer,
		writer:            config.Writer,
		dataCh:            make(chan T, config.ChannelBuffer),
		batchCh:           make(chan B, config.ChannelBuffer),
		doneCh:            make(chan struct{}),
		saveBatchesCh:     config.SaveBatchesCh,
	}

	if b.saveBatchesCh == nil {
		b.saveBatchesCh = make(chan struct{}, 1)
	}

	// This slice is needed to broadcast signal from the initial channel to every worker
	manualWriteWorkerChs := make([]chan struct{}, 0, config.Workers)

	b.workersBatchingWg.Add(config.Workers)
	for i := 0; i < config.Workers; i++ {
		manualWriteWorkerCh := make(chan struct{}, 1)
		go b.workerBatch(manualWriteWorkerCh)

		manualWriteWorkerChs = append(manualWriteWorkerChs, manualWriteWorkerCh)
	}

	go broadcastToAll[struct{}](b.saveBatchesCh, manualWriteWorkerChs...)

	b.workersWritingWg.Add(config.Workers)
	for i := 0; i < config.Workers; i++ {
		go b.workerWrite()
	}

	return b
}

// Write saves data to internal batch for further writing. Internally it is writing to channel, so it can be blocked
// for a long time if writing batches is stopped for any reason. Consider passing a copy of data if data can be
// modified outside.
func (l *BatchedWriter[B, T]) Write(data T) {
	l.dataCh <- data
}

// WriteCtx is the same as Write but supports context and may be canceled by cancelling ctx. Consider passing a copy
// of data if data can be modified outside.
func (l *BatchedWriter[B, T]) WriteCtx(ctx context.Context, data T) error {
	select {
	case l.dataCh <- data:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *BatchedWriter[B, T]) SaveBatches() {
	l.saveBatchesCh <- struct{}{}
}

// Close gracefully stops all workers and writing remained data. Blocks until everything is completely written
func (l *BatchedWriter[B, T]) Close() {
	close(l.doneCh)
	close(l.dataCh)
	l.workersBatchingWg.Wait()

	close(l.batchCh)
	l.workersWritingWg.Wait()

	l.writer.Close()
}

// workerBatch is reading from the data channel, handling batch appending and writing to a batch channel
func (l *BatchedWriter[B, T]) workerBatch(manualWritingCh <-chan struct{}) {
	defer l.workersBatchingWg.Done()

	batch := l.batchProducer.NewBatch()

	for {
		select {
		case data, ok := <-l.dataCh:
			if !ok {
				l.batchCh <- batch.Extract()
				return
			}
			if batch.IsFull() {
				l.batchCh <- batch.Extract()
			}
			batch.Append(data)
		case <-l.doneCh:
			for data := range l.dataCh {
				if batch.IsFull() {
					l.batchCh <- batch.Extract()
				}
				batch.Append(data)
			}
			l.batchCh <- batch.Extract()
			return
		case <-manualWritingCh:
			l.batchCh <- batch.Extract()
		}
	}
}

// workerWrite is reading from the batch channel, handling batch writing and resetting internal Writer
func (l *BatchedWriter[B, T]) workerWrite() {
	defer l.workersWritingWg.Done()

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
