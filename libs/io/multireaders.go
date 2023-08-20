package io

import (
	"io"
)

type MultiReaders struct {
	// logger telemetry.Logger
}

func (m *MultiReaders) GenerateMultiReaders(reader io.Reader, count int) []io.Reader {
	readers := make([]io.Reader, count)
	writers := make([]io.Writer, count)

	i := 0
	for i < count {
		r, w := io.Pipe()
		readers[i] = r
		writers[i] = w
		i++
	}

	writer := io.MultiWriter(writers...)
	go func() {
		_, err := io.Copy(writer, reader)
		if err != nil {
			// log error

			for _, w := range writers {
				w.(*io.PipeWriter).CloseWithError(err)
			}
		}

		for _, w := range writers {
			w.(*io.PipeWriter).Close()
		}
	}()

	return readers
}

func NewMultiReaders( /*logger telemetry.Logger*/ ) *MultiReaders {
	return &MultiReaders{ /*logger: logger*/ }
}
