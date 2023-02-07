package alslgr

import (
	"context"
	"time"
)

type (
	Logger interface {
		Write(message []byte) (int, error)

		DumpBuffer() error
		AutoDumpBuffer(interval time.Duration) (<-chan error, context.CancelFunc)
	}

	Dumper interface {
		Dump([]byte) error
	}
)
