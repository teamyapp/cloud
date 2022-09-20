package storage

import (
	"os"
	"path"
	"path/filepath"
)

type FileSystem struct {
	rootDir string
}

var _ MapBackend = (*FileSystem)(nil)

func (f FileSystem) Get(key string) ([]byte, error) {
	return os.ReadFile(path.Join(f.rootDir, key))
}

func (f FileSystem) Put(key string, data []byte) error {
	filePath := path.Join(f.rootDir, key)
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, os.ModePerm)
}

func (f FileSystem) Delete(key string) error {
	return os.RemoveAll(path.Join(f.rootDir, key))
}

func NewFileSystem(rootDir string) FileSystem {
	return FileSystem{rootDir: rootDir}
}
