package storage

import "os"

// FileSystem abstracts file system operations for testability.
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Stat(path string) (os.FileInfo, error)
	Remove(path string) error
	ReadDir(path string) ([]os.DirEntry, error)
	CreateTemp(dir, pattern string) (*os.File, error)
	Rename(oldpath, newpath string) error
}

// RealFileSystem implements FileSystem using the actual OS file system.
type RealFileSystem struct{}

func (RealFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (RealFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (RealFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (RealFileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (RealFileSystem) Remove(path string) error {
	return os.Remove(path)
}

func (RealFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

func (RealFileSystem) CreateTemp(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}

func (RealFileSystem) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}
