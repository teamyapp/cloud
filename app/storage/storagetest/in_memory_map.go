package storagetest

import (
	"fmt"

	"github.com/teamyapp/cloud/app/storage"
	"github.com/teamyapp/cloud/libs/errs"
)

type InMemoryMap struct {
	data map[string][]byte
}

var _ storage.MapBackend = (*InMemoryMap)(nil)

func (i InMemoryMap) Get(key string) ([]byte, *errs.Error) {
	value, ok := i.data[key]
	if !ok {
		return nil, errs.NewError(errs.NotFound, fmt.Sprintf("key not found: key=%v", key))
	}

	return value, nil
}

func (i InMemoryMap) Put(key string, data []byte) *errs.Error {
	i.data[key] = data
	return nil
}

func (i InMemoryMap) Delete(key string) *errs.Error {
	delete(i.data, key)
	return nil
}

func NewInMemoryMap() InMemoryMap {
	return InMemoryMap{data: map[string][]byte{}}
}
