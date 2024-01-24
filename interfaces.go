package alslgr

type (
	// Writer is an abstract writer that is able to write either raw data or aggregated Batch. Note that in general
	// case, its implementation must allow safe calls of any methods from different goroutines. This requirement can be
	// omitted only when you use BatchedWriter with only one worker. Any errors must be handled internally.
	// Writer is typed with B and T, where T is data to write, and B is a batch of T's.
	Writer[B, T any] interface {
		// Reset is needed for resetting internal state, for example, reopening the file, reconnect to something, etc
		Reset()

		// WriteBatch is needed for writing aggregated data
		WriteBatch(batch B)

		// Write is needed to allow writing non-aggregated data. This method may be called only when Close is called
		Write(data T)
	}

	// Batch is a representation of data aggregator. No methods of Batch will be called from different goroutines, so
	// you don't need to carry about it, but it is required for Extract to return copy of underlying data. General
	// flow is - calling Append with new data until IsFull will return true, then call Extract to retrieve copy
	// of batched data which will be written by any of the workers of BatchedWriter.
	// Batch is typed with B and T, where T is data to write, and B is a batch of T's. You can find implemented
	// slice batching in Slice and SliceProducer structs
	Batch[B, T any] interface {
		// IsFull should indicate when batch is ready to write. Please note that time-based batch writing
		// (for example, writing batches every N seconds no matter is it full or not) is recommended to implement via
		// SaveBatchesCh of BatchedWriterConfig, but not in Batch. This approach will more effectively
		// use available resources. See Slice.IsFull for example
		IsFull() bool

		// Append should save data in the underlying collection. See Slice.Append for example
		Append(data T)

		// Extract should retrieve copy of the underlying batch and reset the underlying collection. It is required
		// to return copy because there are no guaranties that B will be written by Writer.WriteBatch before
		// the next Append call. See Slice.Extract for example
		Extract() B
	}

	// BatchProducer is used for creating new Batch's. Method NewBatch will be called only once per BatchedWriter
	// worker, and in case of multiple workers must allow safe calls from different goroutines
	BatchProducer[B, T any] interface {
		//NewBatch must return new Batch
		NewBatch() Batch[B, T]
	}
)
