package alslgr

type (
	// Dumper interface is an abstraction of some storage, where buffered data will be dumped
	Dumper interface {
		// Dump dumps b to the underlying storage.
		Dump(b []byte)
	}
)
