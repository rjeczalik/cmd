package fs

import "os"

// File TODO
type File interface {
	Close() error
	Read([]byte) (int, error)
	Readdir(int) ([]os.FileInfo, error)
	Seek(int64, int) (int64, error)
	Stat() (os.FileInfo, error)
	Write([]byte) (int, error)
}

// Filesystem TODO
type Filesystem interface {
	Create(string) (File, error)
	Mkdir(string, os.FileMode) error
	Open(string) (File, error)
	Remove(string) error
	Stat(string) (os.FileInfo, error)
}

type FS struct{}

func (FS) Create(name string) (File, error) {
	return os.Create(name)
}

func (FS) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (FS) Open(name string) (File, error) {
	return os.Open(name)
}

func (FS) Remove(name string) error {
	return os.Remove(name)
}

func (FS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

var DefaultFilesystem Filesystem = FS{}
