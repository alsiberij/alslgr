package alslgr

import (
	"os"
)

type (
	fileDumper struct {
		dumpFilenameFunc func() string
		filePerms        os.FileMode
	}
)

const (
	FileDumperDefaultPerms = os.FileMode(0640) //rw-r-----
)

func NewFileDumper(dumpFilenameFunc func() string, perms os.FileMode) Dumper {
	return &fileDumper{
		dumpFilenameFunc: dumpFilenameFunc,
		filePerms:        perms,
	}
}

func (d *fileDumper) Dump(b []byte) error {
	f, err := os.OpenFile(d.dumpFilenameFunc(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, d.filePerms)
	if err != nil {
		return err
	}

	_, err = f.Write(b)
	_ = f.Close()
	return err
}
