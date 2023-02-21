package alslgr

import (
	"os"
)

type (
	stdOutDumper struct {
	}
)

// NewStdoutDumper provides an implementation of Dumper interface that is used for dumping data to a standard output
func NewStdoutDumper() Dumper {
	return &stdOutDumper{}
}

func (d *stdOutDumper) Dump(bytes []byte) error {
	_, err := os.Stdout.Write(bytes)
	return err
}
