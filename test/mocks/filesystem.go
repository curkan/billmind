package mocks

import (
	"os"
	"path/filepath"
)

// MockFileSystem implements storage.FileSystem using real OS operations
// scoped to a base directory (typically t.TempDir()).
type MockFileSystem struct {
	BasePath string
}

// NewMockFileSystem creates a new MockFileSystem rooted at basePath.
func NewMockFileSystem(basePath string) *MockFileSystem {
	return &MockFileSystem{BasePath: basePath}
}

// resolve returns the full path within the mock base directory.
func (m *MockFileSystem) resolve(path string) string {
	rel, err := filepath.Rel(m.BasePath, path)
	if err != nil || len(rel) > 1 && rel[:2] == ".." {
		// Path is not within BasePath; use it as a relative subpath.
		return filepath.Join(m.BasePath, path)
	}
	return path
}

func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(m.resolve(path))
}

func (m *MockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(m.resolve(path), data, perm)
}

func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(m.resolve(path), perm)
}

func (m *MockFileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(m.resolve(path))
}

func (m *MockFileSystem) Remove(path string) error {
	return os.Remove(m.resolve(path))
}

func (m *MockFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(m.resolve(path))
}

func (m *MockFileSystem) CreateTemp(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(m.resolve(dir), pattern)
}

func (m *MockFileSystem) Rename(oldpath, newpath string) error {
	return os.Rename(m.resolve(oldpath), m.resolve(newpath))
}
