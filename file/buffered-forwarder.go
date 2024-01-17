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
		TimedForwardingDoneCh <-chan struct{}
		ReopenForwarderDoneCh <-chan struct{}
		ChannelsBuffer        int
	}
)

func NewBufferedForwarder(config Config) alslgr.BufferedForwarded[[][]byte, []byte] {
	dataBatchProducer := NewDataBatchProducer(config.BatchMaxLen)

	dataForwarder := NewDataForwarder(config.Filename, config.LastResortWriter, config.MaxForwarderBufferLen)

	manualForwardingSignalCh := make(chan struct{}, 1)
	go timedForward(config.TimedForwardingDoneCh, config.TimedForwardInterval, manualForwardingSignalCh)

	reopenForwarderCh := make(chan struct{}, 1)
	go reopenForwarderOnSighup(config.ReopenForwarderDoneCh, config.SighupCh, reopenForwarderCh)

	return alslgr.NewBufferedForwarder[[][]byte, []byte](alslgr.Config[[][]byte, []byte]{
		DataBatchProducer:        &dataBatchProducer,
		DataForwarder:            &dataForwarder,
		ManualForwardingSignalCh: manualForwardingSignalCh,
		ReopenForwarderCh:        reopenForwarderCh,
		ChannelsBuffer:           config.ChannelsBuffer,
		BatchingConcurrency:      1,
		ForwardingConcurrency:    1,
	})
}

func timedForward(doneCh <-chan struct{}, interval time.Duration, signalCh chan<- struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-doneCh:
			return
		case <-ticker.C:
			signalCh <- struct{}{}
		}
	}
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
