package alslgr

import (
	"os"
)

type (
	stdOutDumper struct {
	}
)

func NewStdOutDumper() Dumper {
	return &stdOutDumper{}
}

func (d *stdOutDumper) Dump(bytes []byte) error {
	_, err := os.Stdout.Write(bytes)
	return err
}
