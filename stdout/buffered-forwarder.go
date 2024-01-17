package stdout

import (
	"github.com/alsiberij/alslgr/v3"
)

type (
	Config struct {
		BatchMaxLen           int
		ChannelsBuffer        int
		BatchingConcurrency   int
		ForwardingConcurrency int
	}
)

func NewBufferedForwarder(config Config) alslgr.BufferedForwarded[[][]byte, []byte] {
	dataBatchProducer := NewDataBatchProducer(config.BatchMaxLen)
	dataForwarder := NewDataForwarder()

	return alslgr.NewBufferedForwarder[[][]byte, []byte](alslgr.Config[[][]byte, []byte]{
		DataBatchProducer:        &dataBatchProducer,
		DataForwarder:            &dataForwarder,
		ManualForwardingSignalCh: nil,
		ReopenForwarderCh:        nil,
		ChannelsBuffer:           config.ChannelsBuffer,
		BatchingConcurrency:      config.BatchingConcurrency,
		ForwardingConcurrency:    config.ForwardingConcurrency,
	})
}
