package file

type (
	// batch implements alslgr.Batch with []byte for B and T. Internally it is just a buffer
	batch []byte
)

// TryAppend tries to save data in internal buffer and returns true if it fits, false otherwise
func (b *batch) TryAppend(data []byte) bool {
	if len(data) > cap(*b)-len(*b) {
		return false
	}

	*b = append(*b, data...)
	return true
}

// Extract returns a copy of internal buffer and resets its length to zero
func (b *batch) Extract() []byte {
	cp := make([]byte, len(*b))
	copy(cp, *b)
	*b = (*b)[:0]
	return cp
}

// Pack returns data itself
func (b *batch) Pack(data []byte) []byte {
	return data
}
