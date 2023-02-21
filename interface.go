package alslgr

import (
	"context"
	"time"
)

type (
	// Logger interface is an abstraction of buffered logger, that is able to write data to its buffer and dump
	// that buffer manually or automatically within a time interval to an abstract Dumper
	Logger interface {
		// Write writes b to internal buffer. If there is no free space, buffer will be dumped. If dumping
		// was not successful, Write will return corresponded error
		Write(b []byte) (int, error)

		// DumpBuffer performs manual dump of internal buffer
		DumpBuffer() error

		// AutoDumpBuffer starts automatic dumping of internal buffer. Errors that occurred while dumping
		// can be read from returned channel. Keep in mind that automatic dumping will be blocked until you read
		// a value from returned channel (even nil).
		// Automatic dumping can be stopped via calling returned context.CancelFunc
		AutoDumpBuffer(interval time.Duration) (<-chan error, context.CancelFunc)
	}

	// Dumper interface is an abstraction of some storage, where buffered data will be dumped
	Dumper interface {
		// Dump dumps b to the underlying storage. If entire b was successfully dumped, nil will be returned
		Dump(b []byte) error
	}
)
