package storage

import (
	"os"
	"path"
	"path/filepath"

	"github.com/teamyapp/cloud/libs/errs"
)

type FileSystem struct {
	rootDir string
}

var _ MapBackend = (*FileSystem)(nil)

func (f FileSystem) Get(key string) ([]byte, *errs.Error) {
	buf, err := os.ReadFile(path.Join(f.rootDir, key))
	if err != nil {
		return nil, &errs.Error{
			Code:     errs.OS,
			EmbedErr: err,
		}
	}

	return buf, nil
}

func (f FileSystem) Put(key string, data []byte) *errs.Error {
	filePath := path.Join(f.rootDir, key)
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return &errs.Error{
			Code:     errs.OS,
			EmbedErr: err,
		}
	}

	err = os.WriteFile(filePath, data, os.ModePerm)
	if err != nil {
		return &errs.Error{
			Code:     errs.OS,
			EmbedErr: err,
		}
	}

	return nil
}

func (f FileSystem) Delete(key string) *errs.Error {
	err := os.RemoveAll(path.Join(f.rootDir, key))
	if err != nil {
		return &errs.Error{
			Code:     errs.OS,
			EmbedErr: err,
		}
	}

	return nil
}

func NewFileSystem(rootDir string) FileSystem {
	return FileSystem{rootDir: rootDir}
}
