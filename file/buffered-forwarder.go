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
		BatchMaxLen             int
		MaxForwarderBufferLen   int
		Filename                string
		LastResortWriter        io.Writer
		ChannelsBuffer          int
		TimedForwardingInterval time.Duration
		TimedForwardingDoneCh   <-chan struct{}
		SighupCh                <-chan os.Signal
		SigHupListeningDoneCh   <-chan struct{}
	}
)

func NewBufferedForwarder(config Config) alslgr.BufferedForwarder[[][]byte, []byte] {
	bp := alslgr.SliceBatchProducer[[]byte](config.BatchMaxLen)
	fwd := newForwarder(config.Filename, config.LastResortWriter, config.MaxForwarderBufferLen)

	reopenForwarderCh := make(chan struct{}, 1)
	go reopenFileOnSighup(config.SigHupListeningDoneCh, config.SighupCh, reopenForwarderCh)

	var manualForwardingCh <-chan struct{}
	if config.TimedForwardingInterval > 0 {
		manualForwardingCh = alslgr.NewTicker(config.TimedForwardingInterval, config.TimedForwardingDoneCh)
	}

	return alslgr.NewBufferedForwarder[[][]byte, []byte](alslgr.Config[[][]byte, []byte]{
		BatchProducer:         &bp,
		Forwarder:             &fwd,
		ManualForwardingCh:    manualForwardingCh,
		ResetForwarderCh:      reopenForwarderCh,
		ChannelsBuffer:        config.ChannelsBuffer,
		BatchingConcurrency:   1,
		ForwardingConcurrency: 1,
	})
}

func reopenFileOnSighup(doneCh <-chan struct{}, sigHupCh <-chan os.Signal, reopenForwarderCh chan<- struct{}) {
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
