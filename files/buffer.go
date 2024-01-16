package files

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
		WriterErrorsHandler   WriterErrorsHandler
		LastResortWriter      io.WriteCloser
		MaxForwarderBufferLen int
		TimedForwardInterval  time.Duration
		SighupCh              <-chan os.Signal
		TimedForwardingDoneCh <-chan struct{}
		ReopenForwarderDoneCh <-chan struct{}
		ChannelsBuffer        int
	}
)

func NewBuffer(config Config) (alslgr.Buffer[[][]byte, []byte], error) {
	batchProducer := NewBatchProducer(config.BatchMaxLen)
	dataForwarder, err := NewForwarder(config.Filename, config.WriterErrorsHandler, config.LastResortWriter, config.MaxForwarderBufferLen)
	if err != nil {
		return alslgr.Buffer[[][]byte, []byte]{}, err
	}

	manualForwardSignalCh := make(chan struct{}, 1)
	go timedForward(config.TimedForwardingDoneCh, config.TimedForwardInterval, manualForwardSignalCh)

	reopenForwarderCh := make(chan struct{}, 1)
	go reopenForwarderOnSighup(config.ReopenForwarderDoneCh, config.SighupCh, reopenForwarderCh)

	return alslgr.NewBuffer[[][]byte, []byte](&batchProducer, &dataForwarder, manualForwardSignalCh, reopenForwarderCh,
		config.ChannelsBuffer, 1, 1), nil
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
