package lang

import (
	"io"
	"time"
)

type Runtime struct {
	Now                 func() time.Time
	Output              io.Writer
	CustomNativeGlobals map[string]any
}

func DefaultRuntime() *Runtime {
	return &Runtime{
		Now: time.Now,
	}
}
