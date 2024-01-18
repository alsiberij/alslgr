package file

import (
	"github.com/alsiberij/alslgr/v3"
	"io"
	"os"
	"syscall"
	"time"
)

type (
	Config struct {
		BatchMaxLen           int
		Filename              string
		LastResortWriter      io.Writer
		MaxForwarderBufferLen int
		TimedForwardInterval  time.Duration
		SighupCh              <-chan os.Signal
		SigHupListeningDoneCh <-chan struct{}
		TimedForwardingDoneCh <-chan struct{}
		ChannelsBuffer        int
	}
)

func NewBufferedForwarder(config Config) alslgr.BufferedForwarder[[][]byte, []byte] {
	dataBatchProducer := NewDataBatchProducer(config.BatchMaxLen)

	dataForwarder := NewDataForwarder(config.Filename, config.LastResortWriter, config.MaxForwarderBufferLen)

	reopenForwarderCh := make(chan struct{}, 1)
	go reopenForwarderOnSighup(config.SigHupListeningDoneCh, config.SighupCh, reopenForwarderCh)

	return alslgr.NewBufferedForwarder[[][]byte, []byte](alslgr.Config[[][]byte, []byte]{
		DataBatchProducer:        &dataBatchProducer,
		DataForwarder:            &dataForwarder,
		ManualForwardingSignalCh: alslgr.TimedForwarding(config.TimedForwardInterval, config.TimedForwardingDoneCh),
		ReopenForwarderCh:        reopenForwarderCh,
		ChannelsBuffer:           config.ChannelsBuffer,
		BatchingConcurrency:      1,
		ForwardingConcurrency:    1,
	})
}

func reopenForwarderOnSighup(doneCh <-chan struct{}, sigHupCh <-chan os.Signal, reopenForwarderCh chan<- struct{}) {
	for {
		select {
		case <-doneCh:
			return
		case signal := <-sigHupCh:
			if signal == syscall.SIGHUP {
				reopenForwarderCh <- struct{}{}
			}
		}
	}
}
