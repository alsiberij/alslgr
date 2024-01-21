package alslgr

type (
	// Writer is an abstract writer that is able to write either raw data or aggregated Batch. Note that in general
	// case, its implementation must allow safe calls of any methods from different goroutines. This requirement can be
	// omitted only when you use BatchedWriter with only one write worker. Any errors must be handled internally.
	Writer[B, T any] interface {
		// Reset is needed for such purposes as resetting underlying connections, reopening underlying file, etc.
		Reset()

		// WriteBatch is needed to allow writing aggregated data. Then method will be called in most cases.
		WriteBatch(batch B)

		// Write is needed to allow writing non-aggregated data.
		Write(data T)

		// Close is needed to gracefully stop Writer and handle closing of any underlying entities
		Close()
	}

	// Batch is a representation of data aggregator. No methods of Batch will be called from different goroutines, so
	// you don't need to carry about it, but it is required for Extract to return copy of underlying data. General
	// flow is - call Append with new data until IsFull will return true, then call Extract to retrieve copy
	// of batched data which will be written by any of the writing workers of BatchedWriter
	Batch[B, T any] interface {
		// IsFull method should indicate when batch is ready to write. Please note that time-based batch writing
		// (for example, writing batches every N seconds no matter is it full or not) is recommended to implement via
		// manualWritingChannel of BatchedWriter, but not in Batch. This approach will more effectively use available
		// resources.
		IsFull() bool

		// Append is needed for saving data in the underlying collection.
		Append(data T)

		// Extract is needed for retrieving copy underlying batch of data. It is required to return copy because
		// there are no guaranties that data will be written by Writer.WriteBatch before the next Append call
		Extract() B
	}

	// BatchProducer is used for creating new Batch's. Method NewBatch will be called only once per batching worker
	// of BatchedWriter and must allow safe calls from different goroutines
	BatchProducer[B, T any] interface {
		//NewBatch must return new Batch
		NewBatch() Batch[B, T]
	}
)
