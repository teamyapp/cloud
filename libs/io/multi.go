package io

import (
	"io"
)

type MultiWriteCloser struct {
	writeClosers []io.WriteCloser
}

func (m MultiWriteCloser) Write(p []byte) (n int, err error) {
	for _, writeCloser := range m.writeClosers {
		n, err = writeCloser.Write(p)
		if err != nil {
			return n, err
		}
	}

	return len(p), nil
}

func (m MultiWriteCloser) Close() error {
	for _, writeCloser := range m.writeClosers {
		err := writeCloser.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

var _ io.WriteCloser = (*MultiWriteCloser)(nil)

func NewMultiWriteCloser(writeClosers ...io.WriteCloser) MultiWriteCloser {
	return MultiWriteCloser{
		writeClosers: writeClosers,
	}
}
