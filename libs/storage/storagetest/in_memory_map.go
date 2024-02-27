package storagetest

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/storage"
)

type InMemoryMap struct {
	data map[string][]byte
}

var _ storage.ObjectStore = (*InMemoryMap)(nil)

func (*InMemoryMap) GetDataStreams(ct context.Context, key string) ([]storage.ObjectDataStream, *errs.Error) {
	panic("unimplemented")
}

func (*InMemoryMap) GetMetadata(ct context.Context, key string) (storage.ObjectMetadata, *errs.Error) {
	panic("unimplemented")
}

func (i *InMemoryMap) Get(ct context.Context, key string) (io.Reader, *errs.Error) {
	value, ok := i.data[key]
	if !ok {
		return nil, errs.NewError(errs.NotFound, fmt.Sprintf("key not found: key=%v", key))
	}

	return bytes.NewReader(value), nil
}

func (i *InMemoryMap) Put(ct context.Context, key string, data io.Reader, objectMetadataInput storage.ObjectMetadataInput) *errs.Error {
	reader, err := io.ReadAll(data)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	i.data[key] = reader
	return nil
}

func (i *InMemoryMap) Delete(ct context.Context, key string) *errs.Error {
	delete(i.data, key)
	return nil
}

func NewInMemoryMap() *InMemoryMap {
	return &InMemoryMap{data: map[string][]byte{}}
}
