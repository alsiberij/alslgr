package alslgr

type (
	DataForwarder[B, T any] interface {
		Reopen()

		ForwardDataBatch(batch B)
		ForwardData(data T)

		Close()
	}

	DataBatch[B, T any] interface {
		ReadyToSend() bool
		Append(data T)
		Extract() B
	}

	DataBatchProducer[B, T any] interface {
		NewDataBatch() DataBatch[B, T]
	}
)
