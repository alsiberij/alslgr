package file

import (
	"github.com/alsiberij/alslgr/v3"
	"io"
)

// NewBatchedWriter returns alslgr.BatchedWriter with writing data in a file. This instance is typed
// with []byte and also uses []byte for batching. Keep in mind that this instance uses only one worker in order to
// achieve consequential writing.
// The batchMaxLen indicates how many entries should be aggregated in a batch. The maxBufferLenBeforeWriting is used
// to set the maximum length of the internal buffer which will be used before writing in file. Using it reduces
// amount of system calls and significantly improves performance. General formula for it is average size of entry
// multiplied by BatchMaxLen. The filename is the name of file which will be used for writing. The lastResortWriter
// will be used in case of any error that my occur while writing data in file. Errors of LastResortWriter are
// ignored. Be careful passing nil here, in case of file write error panic will occur
func NewBatchedWriter(maxBufferLength int, filename string,
	lastResortWriter io.Writer) (bw alslgr.BatchedWriter[[]byte, []byte], err error) {

	bp := batchProducer(maxBufferLength)

	w, err := newWriter(filename, lastResortWriter)
	if err != nil {
		return bw, err
	}

	return alslgr.NewBatchedWriter[[]byte, []byte](&bp, &w, 1), nil
}
