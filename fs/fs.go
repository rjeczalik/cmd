// Package fs provides an interface for a filesystem.
package fs

import (
	"os"
	"path/filepath"
)

// File is an almost complete interface for the *os.File.
type File interface {
	// Close closes the underlying file.
	// Default implementation wraps (*os.File).Close.
	Close() error
	// Read reads the content of the File.
	// Default implementation wraps (*os.File).Read.
	Read([]byte) (int, error)
	// Readdir gives the file list if the File is a directory.
	// Default implementation wraps (*os.File).Readdir.
	Readdir(int) ([]os.FileInfo, error)
	// Seek moves the current file offset.
	// Default implementation wraps (*os.File).Seek.
	Seek(int64, int) (int64, error)
	// Stat gives the File details.
	// Default implementation wraps (*os.File).Stat.
	Stat() (os.FileInfo, error)
	// Write writes data to the File.
	// Default implementation wraps (*os.File).Write.
	Write([]byte) (int, error)
}

// Filesystem provides an interface for operating on named files.
type Filesystem interface {
	// Create creates new file or truncates existing one.
	Create(string) (File, error)
	// Mkdir creates new directory. It's a nop, if the directory already exists.
	Mkdir(string, os.FileMode) error
	// MkdirAll creates new directory and all its parents, if needed.
	MkdirAll(string, os.FileMode) error
	// Open opens a file or a directory given by the path.
	Open(string) (File, error)
	// Remove deletes a file given by the path.
	Remove(string) error
	// Stat gives a file or a directory details, given by the path.
	Stat(string) (os.FileInfo, error)
	// Walk walks the file tree starting at root, calling WalkFunc for each file
	// or directory.
	Walk(string, filepath.WalkFunc) error
}

// FS provides an implementation for Filesystem interface, wrapping functions
// from the os package.
type FS struct{}

// Create wraps os.Create.
func (FS) Create(name string) (File, error) {
	return os.Create(name)
}

// Mkdir wraps os.Mkdir.
func (FS) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

// Mkdir wraps os.MkdirAll.
func (FS) MkdirAll(name string, perm os.FileMode) error {
	return os.MkdirAll(name, perm)
}

// Open wraps os.Open.
func (FS) Open(name string) (File, error) {
	return os.Open(name)
}

// Remove wraps os.Remove.
func (FS) Remove(name string) error {
	return os.Remove(name)
}

// Stat wraps os.Stat.
func (FS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Walk wraps filepath.Walk.
func (FS) Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}

// Default is the default implementation of Filesystem, which wraps functions
// from the os package.
var Default Filesystem = FS{}
