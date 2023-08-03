# Description

Package defines **Logger** with internal buffer in order to minimize amount of expensive function calls,
e.g. opening/closing file. Also, when these calls are performed, internal buffer if substituting with additional
one, so writing to **Logger** isn't blocked while old buffer being dumped.
Package also defines **Dumper** interface, that allows you to dump buffered data wherever you want.

# Example

```go
package main

import (
	"github.com/alsiberij/alslgr/v2"
	"os"
)

type (
	stdOutDumper struct {}
)

func (d *stdOutDumper) Dump(b []byte) {
	_, _ = os.Stdout.Write(b)
}

func main() {
	dumper := &stdOutDumper{}

	logger := alslgr.NewLogger(100_000, dumper)
	logger.Start()
	defer logger.Stop()
	
	//logger.StartAutoDump(time.Second)

	logger.Write([]byte("Hello, world!"))
}
```