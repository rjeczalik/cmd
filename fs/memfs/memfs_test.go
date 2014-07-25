package memfs

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var small = []byte(".\nfs\n\tfs.go\n\tmemfs\n\t\tmemfs.go\n\t\tmemfs_test.go\n" +
	"LICENSE\nREADME.md\n")

var large = []byte(".\na\n\tb1\n\t\tc1\n\t\t\tc1.txt\n\t\tc2\n\t\t\tc2.txt\n\t\t" +
	"c3\n\t\t\tc3.txt\n\t\t\td1\n\t\t\t\te1\n\t\t\t\t\t_\n\t\t\t\t\t\t_.txt" +
	"\n\t\t\t\t\te1.txt\n\t\t\t\t\te2.txt\n\t\t\t\t\te/\n\tb2\n\t\tc1\n\t\t" +
	"\td1.txt\n\t\t\td2/\n\t\t\td3.txt\na.txt\nw\n\tw.txt\n\tx\n\t\ty\n\t\t" +
	"\tz\n\t\t\t\t1.txt\n\t\ty.txt\n")

func TestCreate(t *testing.T) {
	fs := Must(TabTree(small))
	cases := [...]struct {
		file string
		err  error
	}{
		0:  {file: "c:/fs/memfs/all_test.go"},
		1:  {file: "/LICENSE"},
		2:  {file: "c:/fs/fs.go"},
		3:  {file: "/LICENSE.md"},
		4:  {file: "/fs/fs_test.go"},
		5:  {file: "/", err: (*os.PathError)(nil)},
		6:  {file: "c:", err: (*os.PathError)(nil)},
		7:  {file: "c:/", err: (*os.PathError)(nil)},
		8:  {file: "/fs", err: (*os.PathError)(nil)},
		9:  {file: "/fs/memfs", err: (*os.PathError)(nil)},
		10: {file: "/.git/config", err: (*os.PathError)(nil)},
		11: {file: "/fs/.svn/config", err: (*os.PathError)(nil)},
		12: {file: "/LICENSE/OTHER.md", err: (*os.PathError)(nil)},
		13: {file: "/fs/fs.go/detail.go", err: (*os.PathError)(nil)},
		14: {file: "/fs/memfs/nfs/nfs.go", err: (*os.PathError)(nil)},
	}
	for i, cas := range cases {
		file := filepath.FromSlash(cas.file)
		f, err := fs.Create(file)
		if cas.err == nil && err != nil {
			t.Errorf("want err=nil; was %q (i=%d)", err, i)
			continue
		}
		if cas.err != nil && err == nil {
			t.Errorf("want typeof(err)=%T; was nil (i=%d)", cas.err, i)
			continue
		}
		if cas.err != nil && err != nil {
			if reflect.TypeOf(cas.err) != reflect.TypeOf(err) {
				t.Errorf("want typeof(err)=%T; was %T (i=%d)", cas.err, err, i)
			}
			continue
		}
		fi, err := f.Stat()
		if err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if fi.Name() != file {
			t.Errorf("want fi.Name()=%q; got %q (i=%d)", file, fi.Name(), i)
		}
		if fi.IsDir() {
			t.Errorf("want fi.IsDir()=false; got true (i=%d)", i)
		}
	}
}

func TestMkdir(t *testing.T) {
	fs := Must(TabTree(small))
	cases := [...]struct {
		dir string
		err error
	}{
		0:  {dir: "/testdata"},
		1:  {dir: "/fs/testdata"},
		2:  {dir: "c:/fs/memfs/testdata"},
		3:  {dir: "c:/testdata"},
		4:  {dir: "c:/"},
		5:  {dir: "/"},
		6:  {dir: "c:/LICENSE", err: (*os.PathError)(nil)},
		7:  {dir: "c:/LICENSE/testdata", err: (*os.PathError)(nil)},
		8:  {dir: "/fs/memfs/memfs.go", err: (*os.PathError)(nil)},
		9:  {dir: "/fs/fs.go/testdata", err: (*os.PathError)(nil)},
		10: {dir: "c:/fs/memfs/memfs_test.go", err: (*os.PathError)(nil)},
	}
	for i, cas := range cases {
		dir := filepath.FromSlash(cas.dir)
		err := fs.Mkdir(dir, 0xD)
		if cas.err == nil && err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if cas.err != nil && err == nil {
			t.Errorf("want typeof(err)=%T; was nil (i=%d)", cas.err, i)
			continue
		}
		if cas.err != nil && err != nil {
			if reflect.TypeOf(cas.err) != reflect.TypeOf(err) {
				t.Errorf("want typeof(err)=%T; was %T (i=%d)", cas.err, err, i)
			}
			continue
		}
		fi, err := fs.Stat(dir)
		if err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if fi.Name() != dir {
			t.Errorf("want fi.Name()=%q; got %q (i=%d)", dir, fi.Name(), i)
		}
		if !fi.IsDir() {
			t.Errorf("want fi.IsDir()=true; got false (i=%d)", i)
		}
	}
}

func TestMkdirNop(t *testing.T) {
	fs := Must(TabTree(large))
	cases := [...]string{
		0: "/a/b1",
		1: "/",
		2: "/w/x/y",
		3: "/a/b1/c3/d1",
		4: "/a/b2/c1",
	}
	for i, mkdir := range []func(FS, string, os.FileMode) error{FS.Mkdir, FS.MkdirAll} {
		for j, cas := range cases {
			dir := filepath.FromSlash(cas)
			mutfs := Must(TabTree(large))
			if err := mkdir(mutfs, dir, 0xD); err != nil {
				t.Errorf("want err=nil; got %q (i=%d, j=%d)", err, i, j)
				continue
			}
			if !Compare(Must(mutfs.Cd(dir)), Must(fs.Cd(dir))) {
				t.Errorf("want Compare(...)=true; got false (i=%d, j=%d)", i, j)
			}
		}
	}
}

func TestMkdirAll(t *testing.T) {
	fs := Must(TabTree(small))
	cases := [...]struct {
		dir string
		err error
	}{
		0:  {dir: "/"},
		1:  {dir: "/abc"},
		2:  {dir: "/abc/1/2/3"},
		3:  {dir: "/fs/abc"},
		4:  {dir: "/fs/abc/1/2/3"},
		5:  {dir: "/fs/memfs/abc/1/2/3"},
		6:  {dir: "/fs/fs.go/testdata", err: (*os.PathError)(nil)},
		7:  {dir: "/LICENSE", err: (*os.PathError)(nil)},
		8:  {dir: "/README.md/testdata", err: (*os.PathError)(nil)},
		9:  {dir: "/fs/memfs/memfs.go/abc", err: (*os.PathError)(nil)},
		10: {dir: "/fs/memfs/memfs.go/abc/1/2/3", err: (*os.PathError)(nil)},
	}
	for i, cas := range cases {
		dir := filepath.FromSlash(cas.dir)
		err := fs.MkdirAll(dir, 0xD)
		if cas.err == nil && err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if cas.err != nil && err == nil {
			t.Errorf("want typeof(err)=%T; was nil (i=%d)", cas.err, i)
			continue
		}
		if cas.err != nil && err != nil {
			if reflect.TypeOf(cas.err) != reflect.TypeOf(err) {
				t.Errorf("want typeof(err)=%T; was %T (i=%d)", cas.err, err, i)
			}
			continue
		}
		fi, err := fs.Stat(dir)
		if err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if !fi.IsDir() {
			t.Errorf("want fi.IsDir()=true; got false (i=%d)", i)
		}
	}
}

func TestOpen(t *testing.T) {
	fs := Must(TabTree(small))
	cases := [...]struct {
		path string
		dir  bool
	}{
		0: {"c:/", true},
		1: {"/", true},
		2: {"/fs", true},
		3: {"c:/fs/memfs", true},
		4: {"/LICENSE", false},
		5: {"c:/README.md", false},
		6: {"/fs/fs.go", false},
		7: {"c:/fs/memfs/memfs.go", false},
		8: {"/fs/memfs/memfs_test.go", false},
	}
	for i, cas := range cases {
		path := filepath.FromSlash(cas.path)
		f, err := fs.Open(path)
		if err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		fi, err := f.Stat()
		if err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if fi.Name() != path {
			t.Errorf("want fi.Name()=%q; got %q (i=%d)", path, fi.Name(), i)
		}
		if fi.IsDir() != cas.dir {
			t.Errorf("want fi.IsDir()=%v; got %v (i=%d)", cas.dir, fi.IsDir(), i)
		}
	}
}

func TestRemove(t *testing.T) {
	fs := Must(TabTree(small))
	cases := [...]struct {
		file string
		err  error
	}{
		0:  {file: "/LICENSE"},
		1:  {file: "/README.md"},
		2:  {file: "/fs", err: (*os.PathError)(nil)},
		3:  {file: "/fs/fs.go"},
		4:  {file: "/fs/memfs", err: (*os.PathError)(nil)},
		5:  {file: "/fs/memfs/memfs.go"},
		6:  {file: "/fs/memfs/memfs_test.go"},
		7:  {file: "/", err: (*os.PathError)(nil)},
		8:  {file: "c:", err: (*os.PathError)(nil)},
		9:  {file: "/er234", err: os.ErrNotExist},
		10: {file: "/fs/dfgdft345", err: os.ErrNotExist},
	}
	for i, cas := range cases {
		file := filepath.FromSlash(cas.file)
		err := fs.Remove(file)
		if cas.err == nil && err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if cas.err != nil && err == nil {
			t.Errorf("want typeof(err)=%T; was nil (i=%d)", cas.err, i)
			continue
		}
		if cas.err != nil && err != nil {
			if !reflect.ValueOf(cas.err).IsNil() && os.IsNotExist(cas.err) {
				if !os.IsNotExist(err) {
					t.Errorf("want os.IsNotExist(%v)=true (i=%d)", err, i)
				}
				continue
			}
			if reflect.TypeOf(cas.err) != reflect.TypeOf(err) {
				t.Errorf("want typeof(err)=%T; was %T (i=%d)", cas.err, err, i)
			}
			continue
		}
		if _, err := fs.Stat(file); !os.IsNotExist(err) {
			t.Errorf("want os.IsNotExist(%v)=true (i=%d)", err, i)
		}
	}
}

func TestReaddir(t *testing.T) {
	fs := Must(TabTree(small))
	cases := map[string][]struct {
		name string
		dir  bool
	}{
		"/": {
			{"fs", true},
			{"LICENSE", false},
			{"README.md", false},
		},
		"/fs": {
			{"fs.go", false},
			{"memfs", true},
		},
		"c:/fs/memfs": {
			{"memfs.go", false},
			{"memfs_test.go", false},
		},
	}
	for path, cas := range cases {
		path = filepath.FromSlash(path)
		dir, err := fs.Open(path)
		if err != nil {
			t.Errorf("want err=nil; got %q (path=%q)", err, path)
			continue
		}
		fi, err := dir.Readdir(0)
		if err != nil {
			t.Errorf("want err=nil; got %q (path=%q)", err, path)
			continue
		}
		if len(fi) != len(cas) {
			t.Errorf("want len(fi)=%d; got %d (path=%q)", len(cas), len(fi), path)
			continue
		}
	LOOP:
		for _, it := range cas {
			s := filepath.Join(path, it.name)
			for _, fi := range fi {
				if fi.Name() == s {
					if fi.IsDir() != it.dir {
						t.Errorf("want fi.IsDir()=%v; got %v (path=%q)", it.dir, fi.IsDir(), s)
					}
					continue LOOP
				}
			}
			t.Errorf("%q not found in fi", path)
		}
	}
}

func TestCd(t *testing.T) {
	fs := Must(TabTree(large))
	cases := [...]struct {
		path string
		fs   []byte
	}{{
		"/a/b1/c3",
		[]byte(".\nc3.txt\nd1\n\te1\n\t\t_\n\t\t\t_.txt\n\t\te1.txt\n\t\te2.txt\n\t\te/"),
	}, {
		"/a/b2",
		[]byte(".\nc1\n\td1.txt\n\td2/\n\td3.txt"),
	}, {
		"/w/x",
		[]byte(".\ny\n\tz\n\t\t1.txt\ny.txt"),
	}, {
		"/a/b1/c3/d1/e1/_",
		[]byte(".\n_.txt"),
	}, {
		"/w/x/y/z",
		[]byte(".\n1.txt"),
	}}
	for i, cas := range cases {
		path := filepath.FromSlash(cas.path)
		rhs := Must(TabTree(cas.fs))
		lhs, err := fs.Cd(path)
		if err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if !Compare(lhs, rhs) {
			t.Errorf("want Compare(...)=true; got false (i=%d)", i)
		}
	}
}
