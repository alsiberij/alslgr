package alslgr

type (
	DataForwarder[B, T any] interface {
		Reopen()

		ForwardBatch(batch B)
		Forward(data T)

		Close()
	}

	Batch[B, T any] interface {
		ReadyToSend() bool
		Append(data T)
		Extract() B
	}

	BatchProducer[B, T any] interface {
		NewBatch() Batch[B, T]
	}
)
