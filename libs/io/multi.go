package io

import (
	"io"
)

type MultiWriteCloser struct {
	writeClosers []io.WriteCloser
}

var _ io.WriteCloser = (*MultiWriteCloser)(nil)

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
	var err error
	for _, writeCloser := range m.writeClosers {
		newErr := writeCloser.Close()
		if err == nil {
			err = newErr
		}
	}

	return err
}

func NewMultiWriteCloser(writeClosers ...io.WriteCloser) MultiWriteCloser {
	return MultiWriteCloser{
		writeClosers: writeClosers,
	}
}
