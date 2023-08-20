package storage

import (
	"bytes"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/teamyapp/cloud/libs/errs"
)

type FileSystem struct {
	rootDir string
}

var _ MapClient = (*FileSystem)(nil)

func (f FileSystem) Get(key string) (io.Reader, *errs.Error) {
	buf, err := os.ReadFile(path.Join(f.rootDir, key))
	if err != nil {
		return nil, errs.NewError(errs.OS, err.Error())
	}

	return bytes.NewReader(buf), nil
}

func (f FileSystem) Put(key string, data io.Reader) *errs.Error {
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

func (f FileSystem) Delete(key string) *errs.Error {
	err := os.RemoveAll(path.Join(f.rootDir, key))
	if err != nil {
		return errs.NewError(errs.OS, err.Error())
	}

	return nil
}

func NewFileSystem(rootDir string) FileSystem {
	return FileSystem{rootDir: rootDir}
}
