package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/teamyapp/cloud/libs/errs"
)

type FileSystem struct {
	rootDir string
}

var _ ObjectStore = (*FileSystem)(nil)

func (*FileSystem) GetDataStreams(ct context.Context, key string) ([]ObjectDataStream, *errs.Error) {
	panic("unimplemented")
}

func (*FileSystem) GetMetadata(ct context.Context, key string) (ObjectMetadata, *errs.Error) {
	panic("unimplemented")
}

func (f *FileSystem) Get(ct context.Context, key string) (io.Reader, *errs.Error) {
	buf, err := os.ReadFile(path.Join(f.rootDir, key))
	if err != nil {
		return nil, errs.NewError(errs.OS, err.Error())
	}

	return bytes.NewReader(buf), nil
}

func (f *FileSystem) Put(ct context.Context, key string, data io.Reader, objectMetadataInput ObjectMetadataInput) *errs.Error {
	filePath := path.Join(f.rootDir, key)
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return errs.NewError(errs.OS, err.Error())
	}

	dataBytes, err := io.ReadAll(data)
	err = os.WriteFile(filePath, dataBytes, os.ModePerm)
	if err != nil {
		return errs.NewError(errs.OS, err.Error())
	}

	return nil
}

func (f *FileSystem) Delete(ct context.Context, key string) *errs.Error {
	err := os.RemoveAll(path.Join(f.rootDir, key))
	if err != nil {
		return errs.NewError(errs.OS, err.Error())
	}

	return nil
}

func NewFileSystem(rootDir string) *FileSystem {
	return &FileSystem{rootDir: rootDir}
}
