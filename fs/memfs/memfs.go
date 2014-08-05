// Package memfs provides an interface for an in-memory filesystem.
package memfs

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rjeczalik/tools/fs"
)

// TODO(rjeczalik): do check Property for directory access and file read/write.

const sep = string(os.PathSeparator)

// Umask defines a default file mode creation mask. When not specified explicitely,
// new files has default permission bits equal to 0666 & ~Umask, and 0777 & ~Umask
// for directories.
var Umask os.FileMode = 0002

// Property defines file mode and modification time values for a File and a Directory.
//
// Example
//
// To change propety of a File, change its embedded Property value:
//
//   var authorized = File{Property{0600, time.Now()}}
//
// To change property of a Directory, define it for a special empty key:
//
//   var dotssh = Directory{"": Property{0700}, "authorized_keys": authorized}
type Property struct {
	Mode    os.FileMode // permission bits
	ModTime time.Time   // last modification time
}

var nilp Property

func readproperty(v interface{}) Property {
	switch v := v.(type) {
	case File:
		if v.Property != nilp {
			return v.Property
		}
		return Property{Mode: 0666 & ^Umask}
	case Directory:
		if p, ok := v[""].(Property); ok {
			return p
		}
		return Property{Mode: 0777 & ^Umask}
	}
	panic(errCorrupted)
}

var (
	errDir       = errors.New("is a directory")
	errNotDir    = errors.New("not a directory")
	errCorrupted = errors.New("tree is corrupted")
)

// Order denotes the access order pattern used by an external operation.
type Order uint8

const (
	OrderLexicalAsc  Order = iota // access items in ascending order
	OrderLexicalDesc              // access items in descending order
)

// Directory represents an in-memory directory. Valid directory has each value
// of a File or Directory type, where a key of such value must be non-empty
// and contain no backward nor forward slashes. An empty key has a special
// treatment - when defined, its value must be of a Property type.
type Directory map[string]interface{}

// Ls lists all the files of d directory in given order. It returns nil if
// the directory is empty.
func (d Directory) Ls(order Order) []string {
	if dirlen(d) == 0 {
		return nil
	}
	s := make([]string, 0, dirlen(d))
	for k := range d {
		// Ignore a Property key.
		if k == "" {
			continue
		}
		s = append(s, k)
	}
	switch order {
	case OrderLexicalAsc:
		sort.StringSlice(s).Sort()
	case OrderLexicalDesc:
		sort.Sort(sort.Reverse(sort.StringSlice(s)))
	default:
		panic("invalid order")
	}
	return s
}

// File represents an in-memory file.
type File struct {
	Property        // file properties
	Content  []byte // file content
}

// FS provides an implementation for Filesystem interface, operating on
// an in-memory file tree.
// TODO(rjeczalik): sync.RWMutex
type FS struct {
	Tree Directory
}

var _ fs.Filesystem = (*FS)(nil)

// New returns an empty filesystem.
func New() FS {
	return FS{
		Tree: Directory{},
	}
}

// Cd gives new filesystem with a root starting at the path of the old filesystem.
func (fs FS) Cd(path string) (FS, error) {
	dir, perr := fs.lookup(path)
	if perr != nil {
		return FS{}, perr
	}
	return FS{Tree: dir}, nil
}

// Create creates an in-memory file under the given path.
func (fs FS) Create(name string) (fs.File, error) {
	dir, base, perr := fs.dirbase(name)
	if perr != nil {
		perr.Op = "Create"
		return nil, perr
	}
	if base == "" {
		return nil, &os.PathError{Op: "Create", Path: name, Err: errDir}
	}
	if v, ok := dir[base]; ok {
		if _, ok = v.(Directory); ok {
			return nil, &os.PathError{Op: "Create", Path: name, Err: errDir}
		}
	}
	dir[base] = File{}
	return file{s: name, f: fs.flushcb(dir, base), r: new(bytes.Reader)}, nil
}

// Mkdir creates an in-memory directory under the given path.
func (fs FS) Mkdir(name string, perm os.FileMode) error {
	dir, base, perr := fs.dirbase(name)
	if perr != nil {
		perr.Op = "Mkdir"
		return perr
	}
	if base == "" {
		return nil
	}
	if v, ok := dir[base]; ok {
		if _, ok = v.(Directory); ok {
			return nil
		}
		return &os.PathError{Op: "Mkdir", Path: name, Err: errNotDir}
	}
	dir[base] = Directory{"": Property{Mode: perm}}
	return nil
}

// MkdirAll creates new in-memory directory and all its parents, if needed.
func (fs FS) MkdirAll(name string, perm os.FileMode) error {
	var (
		dir = fs.Tree
		err error
	)
	fn := func(s string) bool {
		v, ok := dir[s]
		if !ok {
			d := Directory{"": Property{Mode: perm}}
			dir[s], dir = d, d
		} else if dir, ok = v.(Directory); !ok {
			err = &os.PathError{Op: "MkdirAll", Path: name, Err: errNotDir}
			return false
		}
		return true
	}
	fs.dirwalk(name, fn)
	return err
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
		return nil, &os.PathError{Op: "Open", Path: name, Err: os.ErrNotExist}
	}
	switch v := dir[base].(type) {
	case File:
		return file{s: name, f: fs.flushcb(dir, base), r: bytes.NewReader([]byte(v.Content))}, nil
	case Directory:
		return directory{s: name, d: v}, nil
	}
	return nil, &os.PathError{Op: "Open", Path: name, Err: errCorrupted}
}

// Remove removes a file from the tree given by the path.
func (fs FS) Remove(name string) error {
	dir, base, perr := fs.dirbase(name)
	if perr != nil {
		perr.Op = "Remove"
		return perr
	}
	if base == "" {
		return &os.PathError{Op: "Remove", Path: name, Err: os.ErrPermission}
	}
	if _, ok := dir[base]; !ok {
		return &os.PathError{Op: "Remove", Path: name, Err: os.ErrNotExist}
	}
	if _, ok := dir[base].(Directory); ok {
		return &os.PathError{Op: "Remove", Path: name, Err: os.ErrPermission}
	}
	delete(dir, base)
	return nil
}

// RemoveAll removes a file or a directory with all its descendants from the tree
// rooted at the given path.
func (fs FS) RemoveAll(name string) error {
	dir, base, perr := fs.dirbase(name)
	if perr != nil {
		perr.Op = "RemoveAll"
		return perr
	}
	if base == "" {
		return &os.PathError{Op: "Remove", Path: name, Err: os.ErrPermission}
	}
	if _, ok := dir[base]; !ok {
		return &os.PathError{Op: "Remove", Path: name, Err: os.ErrNotExist}
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

// Walk walks the file tree in a depth-first lexical order, calling fn for each
// file or directory.
func (fs FS) Walk(root string, fn filepath.WalkFunc) (err error) {
	dir, perr := fs.lookup(root)
	if perr != nil {
		fn(root, nil, err)
		return perr
	}
	if err = fn(root, fileinfo{readproperty(dir), root, 0, true}, nil); err != nil {
		return
	}
	ifn := func(s string, v interface{}, _ []dirQueue) bool {
		s = filepath.Join(root, s)
		var fi = fileinfo{
			p: readproperty(v),
			s: s,
		}
		switch v := v.(type) {
		case File:
			fi.n, fi.d = int64(len(v.Content)), false
		case Directory:
			fi.n, fi.d = 0, true
		default:
			panic(errCorrupted)
		}
		// TODO(rjeczalik): support filepath.SkipDir
		err = fn(s, fi, nil)
		return err == nil
	}
	dfs(dir, ifn)
	return
}

func (fs FS) dirwalk(p string, fn func(string) bool) {
	if p == "" || p == "." {
		return
	}
	i := strings.Index(p, sep) + 1
	if i == 0 || i == len(p) {
		return
	}
	for i < len(p) {
		j := strings.Index(p[i:], sep)
		if j == -1 {
			j = len(p) - i
		}
		if !fn(p[i : i+j]) {
			return
		}
		i += j + 1
	}
}

func (fs FS) lookup(p string) (dir Directory, perr *os.PathError) {
	dir = fs.Tree
	fn := func(name string) bool {
		v, ok := dir[name]
		if !ok {
			perr = &os.PathError{Err: os.ErrNotExist}
			return false
		}
		if dir, ok = v.(Directory); !ok {
			perr = &os.PathError{Err: errNotDir}
			return false
		}
		return true
	}
	fs.dirwalk(p, fn)
	return
}

func (fs FS) dirbase(p string) (Directory, string, *os.PathError) {
	// TODO(rjeczalik): ignore trailing /
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
	// TODO(rjeczalik): return error if dir[name] is not of a File type
	return func(p []byte) {
		f := dir[name].(File)
		f.Content = p
		dir[name] = f
	}
}

type file struct {
	p Property      // file properties
	s string        // file name
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
	return nil, &os.PathError{Op: "Readdir", Path: f.s, Err: nil}
}

func (f file) Seek(offset int64, whence int) (int64, error) {
	return f.r.Seek(offset, whence)
}

func (f file) Stat() (os.FileInfo, error) {
	return fileinfo{f.p, f.s, int64(f.r.Len()), false}, nil
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
	return 0, &os.PathError{Op: "Read", Path: d.s, Err: nil}
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
		// Ignore special empty key.
		if k == "" {
			continue
		}
		if f, ok := v.(File); ok {
			fi = append(fi, fileinfo{
				readproperty(v),
				filepath.Join(d.s, k),
				int64(len(f.Content)),
				false,
			})
		} else {
			fi = append(fi, fileinfo{
				readproperty(v),
				filepath.Join(d.s, k),
				0,
				true,
			})
		}
	}
	return
}

func (d directory) Seek(int64, int) (int64, error) {
	return 0, &os.PathError{Op: "Seek", Path: d.s, Err: nil}
}

func (d directory) Stat() (os.FileInfo, error) {
	return fileinfo{readproperty(d.d), d.s, 0, true}, nil
}

func (d directory) Write([]byte) (int, error) {
	return 0, &os.PathError{Op: "Write", Path: d.s, Err: nil}
}

type fileinfo struct {
	p Property
	s string
	n int64
	d bool
}

func (fi fileinfo) Name() string       { return fi.s }
func (fi fileinfo) Size() int64        { return fi.n }
func (fi fileinfo) Mode() os.FileMode  { return fi.p.Mode }
func (fi fileinfo) ModTime() time.Time { return fi.p.ModTime }
func (fi fileinfo) IsDir() bool        { return fi.d }
func (fi fileinfo) Sys() interface{}   { return nil }
