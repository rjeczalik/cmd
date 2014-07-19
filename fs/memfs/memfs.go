// Package memfs provides an interface for an in-memory filesystem.
package memfs

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rjeczalik/tools/fs"
)

// Directory represents an in-memory directory
type Directory map[string]interface{}

// File represents an in-memory file.
type File []byte

// FS provides an implementation for Filesystem interface, operating on
// an in-memory file tree.
// TODO(rjeczalik): sync.RWMutex
type FS struct {
	Tree Directory
}

var (
	errDir       = errors.New("is a directory")
	errNotDir    = errors.New("not a directory")
	errCorrupted = errors.New("tree is corrupted")
)

const sep = string(os.PathSeparator)

func (fs FS) lookup(p string) (Directory, *os.PathError) {
	if p == "" || p == "." {
		return fs.Tree, nil
	}
	j := strings.Index(p, sep) + 1
	if j == 0 || j == len(p) {
		return fs.Tree, nil
	}
	dir := fs.Tree
	for j < len(p) {
		i := strings.Index(p[j:], sep)
		if i == -1 {
			i = len(p) - j
		}
		v, ok := dir[p[j:j+i]]
		if !ok {
			return nil, &os.PathError{Err: os.ErrNotExist}
		}
		if dir, ok = v.(Directory); !ok {
			return nil, &os.PathError{Err: errNotDir}
		}
		j += i + 1
	}
	return dir, nil
}

func (fs FS) dirbase(p string) (Directory, string, *os.PathError) {
	i := strings.LastIndex(p, sep)
	if i == -1 {
		return fs.Tree, "", nil
	}
	if i == 0 {
		return fs.Tree, p[1:], nil
	}
	dir, perr := fs.lookup(p[:i])
	if perr != nil {
		perr.Path = p
		return nil, "", perr
	}
	return dir, p[i+1:], nil
}

func (fs FS) flushcb(dir Directory, name string) func([]byte) {
	return func(p []byte) {
		dir[name] = File(p)
	}
}

// Create creates an in-memory file under the given path.
func (fs FS) Create(name string) (fs.File, error) {
	dir, base, perr := fs.dirbase(name)
	if perr != nil {
		perr.Op = "Create"
		return nil, perr
	}
	if base == "" {
		return nil, &os.PathError{"Create", name, errDir}
	}
	if v, ok := dir[base]; ok {
		if _, ok = v.(Directory); ok {
			return nil, &os.PathError{"Create", name, errDir}
		}
	}
	dir[base] = File{}
	return file{s: name, f: fs.flushcb(dir, base), r: new(bytes.Reader)}, nil
}

// Mkdir creates an in-memory directory under the given path.
func (fs FS) Mkdir(name string, _ os.FileMode) error {
	dir, base, perr := fs.dirbase(name)
	if perr != nil {
		perr.Op = "Mkdir"
		return perr
	}
	if base == "" {
		return nil
	}
	if v, ok := dir[base]; ok {
		if _, ok = v.(File); ok {
			return &os.PathError{"Mkdir", name, errNotDir}
		}
	}
	dir[base] = Directory{}
	return nil
}

// Open opens a file or directory given by the path.
func (fs FS) Open(name string) (fs.File, error) {
	dir, base, perr := fs.dirbase(name)
	if perr != nil {
		perr.Op = "Open"
		return nil, perr
	}
	if base == "" {
		return directory{s: name, d: dir}, nil
	}
	if _, ok := dir[base]; !ok {
		return nil, &os.PathError{"Open", name, os.ErrNotExist}
	}
	switch v := dir[base].(type) {
	case File:
		return file{s: name, f: fs.flushcb(dir, base), r: bytes.NewReader([]byte(v))}, nil
	case Directory:
		return directory{s: name, d: v}, nil
	}
	return nil, &os.PathError{"Open", name, errCorrupted}
}

// Remove removes a file from the tree given by the path.
func (fs FS) Remove(name string) error {
	dir, base, perr := fs.dirbase(name)
	if perr != nil {
		perr.Op = "Remove"
		return perr
	}
	if base == "" {
		return &os.PathError{"Remove", name, os.ErrPermission}
	}
	if _, ok := dir[base]; !ok {
		return &os.PathError{"Remove", name, os.ErrNotExist}
	}
	if _, ok := dir[base].(Directory); ok {
		return &os.PathError{"Remove", name, os.ErrPermission}
	}
	delete(dir, base)
	return nil
}

// Stat gives the details of a file or a directory given by the path.
func (fs FS) Stat(name string) (os.FileInfo, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	return f.Stat()
}

type file struct {
	s string        // name
	f func([]byte)  // flush callback
	r *bytes.Reader // for reading (io.Seeker)
	w *bytes.Buffer // for writing - merge?
}

func (f file) Close() (err error) {
	if f.w != nil {
		f.f(f.w.Bytes())
		f.w = nil
	}
	return
}

func (f file) Read(p []byte) (int, error) {
	return f.r.Read(p)
}

func (f file) Readdir(int) ([]os.FileInfo, error) {
	return nil, &os.PathError{"Readdir", f.s, nil}
}

func (f file) Seek(offset int64, whence int) (int64, error) {
	return f.r.Seek(offset, whence)
}

func (f file) Stat() (os.FileInfo, error) {
	return fileinfo{f.s, int64(f.r.Len()), false}, nil
}

func (f file) Write(p []byte) (int, error) {
	if f.w == nil {
		f.w = new(bytes.Buffer)
	}
	return f.w.Write(p)
}

type directory struct {
	s string
	d Directory
}

func (d directory) Close() (err error) {
	return
}

func (d directory) Read(p []byte) (int, error) {
	return 0, &os.PathError{"Read", d.s, nil}
}

// TODO(rjeczalik): make it ordered so it actually works
func (d directory) Readdir(n int) (fi []os.FileInfo, err error) {
	if len(d.d) == 0 {
		return nil, errors.New("Readdir: directory is empty")
	}
	if n > 0 {
		return nil, errors.New("Readdir: not implemented")
	}
	fi = make([]os.FileInfo, 0, len(d.d))
	for k, v := range d.d {
		if f, ok := v.(File); ok {
			fi = append(fi, fileinfo{filepath.Join(d.s, k), int64(len(f)), false})
		} else {
			fi = append(fi, fileinfo{filepath.Join(d.s, k), 0, true})
		}
	}
	return
}

func (d directory) Seek(int64, int) (int64, error) {
	return 0, &os.PathError{"Seek", d.s, nil}
}

func (d directory) Stat() (os.FileInfo, error) {
	return fileinfo{d.s, 0, true}, nil
}

func (d directory) Write([]byte) (int, error) {
	return 0, &os.PathError{"Write", d.s, nil}
}

type fileinfo struct {
	s string
	n int64
	d bool
}

func (fi fileinfo) Name() string       { return fi.s }
func (fi fileinfo) Size() int64        { return fi.n }
func (fi fileinfo) Mode() os.FileMode  { return 0 }
func (fi fileinfo) ModTime() time.Time { return time.Time{} }
func (fi fileinfo) IsDir() bool        { return fi.d }
func (fi fileinfo) Sys() interface{}   { return nil }
