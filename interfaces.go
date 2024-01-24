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

		// Close is needed to gracefully stop Writer and handle closing of any underlying entities
		Close()
	}

	// Batch is a representation of data aggregator. No methods of Batch will be called from different goroutines, so
	// you don't need to carry about it, but it is required for Extract to return copy of underlying data and reset
	// its internal state for reusing resources. Batch is typed with B and T, where T is data to write, and B is
	// a batch of T's. You can find implemented slice batching in Slice and SliceProducer structs
	Batch[B, T any] interface {
		// TryAppend should try to save data in the underlying collection and return whether saving was successful.
		// If not, method Extract will be called in order to try to free space. If second call of TryAppend
		// returns false, method Pack will be used. See Slice.TryAppend for example
		TryAppend(data T) bool

		// Extract should retrieve copy of the underlying batch and reset the underlying collection for further reuse.
		// It is required to return copy because there are no guaranties that B will be written by Writer.WriteBatch
		// before the next TryAppend call. See Slice.Extract for example
		Extract() B

		// Pack should be able to pack single T into a B. In some cases, when you have specific limitations on
		// underlying collection, a single data entry can not be fit into a regular batch, so this method will be used.
		// For example, you have []byte and []byte for B and T (batch is just a large buffer) and limitation for
		// batch that it should be less than 1 kilobyte. But for some reason you want to write a 2 kilobyte data.
		// In this case it should be packed directly into a specific batch in order to not exceed regular limitation.
		Pack(data T) B
	}

	// BatchProducer is used for creating new Batch's. Method NewBatch will be called only once per BatchedWriter
	// worker, and in case of multiple workers must allow safe calls from different goroutines
	BatchProducer[B, T any] interface {
		//NewBatch must return new Batch
		NewBatch() Batch[B, T]
	}
)
