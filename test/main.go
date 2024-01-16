package main

import (
	"github.com/alsiberij/alslgr/v3/files"
	"os"
	"time"
)

func main() {
	b, err := files.NewBuffer(files.Config{
		BatchMaxLen:           2,
		Filename:              "test.log",
		WriterErrorsHandler:   nil,
		LastResortWriter:      os.Stdout,
		MaxForwarderBufferLen: 0,
		TimedForwardInterval:  time.Second,
		SighupCh:              nil,
		TimedForwardingDoneCh: nil,
		ReopenForwarderDoneCh: nil,
		ChannelsBuffer:        1,
	})
	if err != nil {
		panic(err)
	}
	defer b.Close()

	b.Write([]byte("HEEEELLEELL"))
}
