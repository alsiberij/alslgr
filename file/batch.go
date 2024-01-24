package file

type (
	batch []byte
)

func (b *batch) TryAppend(data []byte) bool {
	if len(data) > cap(*b)-len(*b) {
		return false
	}

	*b = append(*b, data...)
	return true
}

func (b *batch) Extract() []byte {
	cp := make([]byte, len(*b))
	copy(cp, *b)
	*b = (*b)[:0]
	return cp
}

func (b *batch) Pack(data []byte) []byte {
	return data
}
