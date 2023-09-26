package io

import (
	"io"
)

func NewMultiReaders(reader io.Reader, count int) []io.Reader {
	readers := make([]io.Reader, count)
	writers := make([]io.Writer, count)

	index := 0
	for index < count {
		r, w := io.Pipe()
		readers[index] = r
		writers[index] = w
		index++
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
