package localfs

import (
	"io"
	"os"
	"path/filepath"
)

// LocalFS provides access to local file system.
type LocalFS struct {
	BasePath string
}

// NewLocalFS returns a storage accessor to basePath of local file system.
func NewLocalFS(basePath string) *LocalFS {
	return &LocalFS{
		BasePath: basePath,
	}
}

// Create returns a io.WriteCloser of a file with the name.
func (localfs *LocalFS) Create(fileName string) (io.WriteCloser, error) {
	fullPath := filepath.Join(localfs.BasePath, fileName)
	if err := os.MkdirAll(filepath.Dir(fullPath), os.FileMode(0644)); err != nil {
		return nil, err
	}
	return os.Create(fullPath)
}

// Open returns io.ReadCloser of the file.
func (localfs *LocalFS) Open(fileName string) (io.ReadCloser, error) {
	fullPath := filepath.Join(localfs.BasePath, fileName)
	return os.Open(fullPath)
}

// Delete deletes the file from the storage.
func (localfs *LocalFS) Delete(fileName string) error {
	fullPath := filepath.Join(localfs.BasePath, fileName)
	return os.Remove(fullPath)
}
