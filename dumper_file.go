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
	// FileDumperDefaultPerms are default file permissions:
	//  rw-r-----
	FileDumperDefaultPerms = os.FileMode(0640)
)

// NewFileDumper provides an implementation of Dumper interface that is used for dumping data to a file.
// Argument dumpFilenameFunc is used to retrieve a filename. Example:
//
//	 dumper := alslgr.NewFileDumper(func() string {
//			return time.Now().Format("logs-02-01-2006.log")
//		}, alslgr.FileDumperDefaultPerms)
//
// Every Dump call on Dumper will open file, write data and close it.
// Write error will be returned, close error will be ignored
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
