package files

import "io"

type (
	WriterErrorsHandler interface {
		HandleError(lastResortWriter io.Writer, b []byte, err error)
	}
)
