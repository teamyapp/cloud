package lang

import (
	"io"
	"time"
)

type Runtime struct {
	Now                   func() time.Time
	Output                io.Writer
	CustomNativeFunctions map[string]Callable
}

func DefaultRuntime() *Runtime {
	return &Runtime{
		Now: time.Now,
	}
}
