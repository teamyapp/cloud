package storagetest

import (
	"bytes"
	"fmt"
	"io"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/storage"
)

type InMemoryMap struct {
	data map[string][]byte
}

var _ storage.MapClient = (*InMemoryMap)(nil)

func (i InMemoryMap) Get(key string) (io.Reader, *errs.Error) {
	value, ok := i.data[key]
	if !ok {
		return nil, errs.NewError(errs.NotFound, fmt.Sprintf("key not found: key=%v", key))
	}

	return bytes.NewReader(value), nil
}

func (i InMemoryMap) Put(key string, data io.Reader) *errs.Error {
	reader, err := io.ReadAll(data)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	i.data[key] = reader
	return nil
}

func (i InMemoryMap) Delete(key string) *errs.Error {
	delete(i.data, key)
	return nil
}

func NewInMemoryMap() InMemoryMap {
	return InMemoryMap{data: map[string][]byte{}}
}
