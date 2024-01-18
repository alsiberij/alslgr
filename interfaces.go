package alslgr

type (
	Forwarder[B, T any] interface {
		Reset()
		ForwardBatch(batch B)
		Forward(data T)
		Close()
	}

	Batch[B, T any] interface {
		IsFull() bool
		Append(data T)
		Extract() B
	}

	BatchProducer[B, T any] interface {
		NewBatch() Batch[B, T]
	}
)
