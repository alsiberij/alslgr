package file

import (
	"github.com/alsiberij/alslgr/v3"
	"io"
	"os"
	"syscall"
	"time"
)

type (
	// Config is a set of required parameters for initializing alslgr.BatchedWriter with file writing
	Config struct {
		// BatchMaxLen indicates how many entries should be aggregated in a batch
		BatchMaxLen int

		// MaxBufferLenBeforeWriting is used to set the maximum length of the internal buffer
		// which will be used before writing in file. Using it reduces amount of system calls and significantly
		// improves performance. General formula for it is average size of entry multiplied by BatchMaxLen
		MaxBufferLenBeforeWriting int

		// Filename is the name of file which will be used for writing
		Filename string

		// LastResortWriter is an io.Writer that will be used in case of any error that my occur while writing data
		// in file. Errors of LastResortWriter are ignored. Be careful passing nil here, in case of file write error
		// panic will occur
		LastResortWriter io.Writer

		// ChannelsBuffer is a size of internal channels buffers. Larger values may positively affect performance
		ChannelsBuffer int

		// TimedForwardingInterval is amount of time that should pass before next automatic data forwarding. Any data
		// that is currently batched will be passed to internal alslgr.Writer
		TimedForwardingInterval time.Duration

		// TimedForwardingDoneCh is used to stop the goroutine that sending signals for automatic forwarding. You can
		// either close it or just send empty struct once. Notice that automatic forwarding will be disabled
		// completely
		TimedForwardingDoneCh <-chan struct{}

		// SighupCh will be used for handling the SIGHUP signal. Receiving it will call Close on a previous file
		// and reopen it again
		SighupCh <-chan os.Signal

		// SigHupDoneCh is used to stop the goroutine that handles SIGHUP signal. You can
		// either close it or just send empty struct once. Notice that handling this signal will be disabled
		// completely
		SigHupDoneCh <-chan struct{}
	}
)

// NewBufferedForwarder returns alslgr.BatchedWriter with forwarding data into a file. This instance is typed
// with []byte and uses default alslgr.SliceProducer (which is typed with [][]byte here) for batching. Also, this
// instance supports periodic automatic forwarding and handles SIGHUP signal for reopening
// underlying file. Keep in mind that this instance uses only one batch worker and only one forward
// worker. That approach is used in order to achieve consequential writing batched data
func NewBufferedForwarder(config Config) alslgr.BatchedWriter[[][]byte, []byte] {
	bp := alslgr.SliceProducer[[]byte](config.BatchMaxLen)

	fwd := newForwarder(config.Filename, config.LastResortWriter, config.MaxBufferLenBeforeWriting)

	reopenForwarderCh := make(chan struct{}, 1)
	go reopenFileOnSighup(config.SigHupDoneCh, config.SighupCh, reopenForwarderCh)

	var manualForwardingCh <-chan struct{}
	if config.TimedForwardingInterval > 0 {
		manualForwardingCh = alslgr.NewTicker(config.TimedForwardingInterval, config.TimedForwardingDoneCh)
	}

	return alslgr.NewBatchedWriter[[][]byte, []byte](alslgr.BatchedWriterConfig[[][]byte, []byte]{
		BatchProducer:   &bp,
		Writer:          &fwd,
		SaveBatchesCh:   manualForwardingCh,
		ResetWriterCh:   reopenForwarderCh,
		ChannelBuffer:   config.ChannelsBuffer,
		BatchingWorkers: 1,
		Workers:         1,
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
