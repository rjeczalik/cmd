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
		0:  {file: filepath.FromSlash("c:/fs/memfs/all_test.go")},
		1:  {file: filepath.FromSlash("/LICENSE")},
		2:  {file: filepath.FromSlash("c:/fs/fs.go")},
		3:  {file: filepath.FromSlash("/LICENSE.md")},
		4:  {file: filepath.FromSlash("/fs/fs_test.go")},
		5:  {file: filepath.FromSlash("/"), err: (*os.PathError)(nil)},
		6:  {file: filepath.FromSlash("c:"), err: (*os.PathError)(nil)},
		7:  {file: filepath.FromSlash("c:/"), err: (*os.PathError)(nil)},
		8:  {file: filepath.FromSlash("/fs"), err: (*os.PathError)(nil)},
		9:  {file: filepath.FromSlash("/fs/memfs"), err: (*os.PathError)(nil)},
		10: {file: filepath.FromSlash("/.git/config"), err: (*os.PathError)(nil)},
		11: {file: filepath.FromSlash("/fs/.svn/config"), err: (*os.PathError)(nil)},
		12: {file: filepath.FromSlash("/LICENSE/OTHER.md"), err: (*os.PathError)(nil)},
		13: {file: filepath.FromSlash("/fs/fs.go/detail.go"), err: (*os.PathError)(nil)},
		14: {file: filepath.FromSlash("/fs/memfs/nfs/nfs.go"), err: (*os.PathError)(nil)},
	}
	for i, cas := range cases {
		f, err := fs.Create(cas.file)
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
		if fi.Name() != cas.file {
			t.Errorf("want fi.Name()=%q; got %q (i=%d)", cas.file, fi.Name(), i)
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
		0:  {dir: filepath.FromSlash("/testdata")},
		1:  {dir: filepath.FromSlash("/fs/testdata")},
		2:  {dir: filepath.FromSlash("c:/fs/memfs/testdata")},
		3:  {dir: filepath.FromSlash("c:/testdata")},
		4:  {dir: filepath.FromSlash("c:/")},
		5:  {dir: filepath.FromSlash("/")},
		6:  {dir: filepath.FromSlash("c:/LICENSE"), err: (*os.PathError)(nil)},
		7:  {dir: filepath.FromSlash("c:/LICENSE/testdata"), err: (*os.PathError)(nil)},
		8:  {dir: filepath.FromSlash("/fs/memfs/memfs.go"), err: (*os.PathError)(nil)},
		9:  {dir: filepath.FromSlash("/fs/fs.go/testdata"), err: (*os.PathError)(nil)},
		10: {dir: filepath.FromSlash("c:/fs/memfs/memfs_test.go"), err: (*os.PathError)(nil)},
	}
	for i, cas := range cases {
		err := fs.Mkdir(cas.dir, 0xD)
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
		fi, err := fs.Stat(cas.dir)
		if err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if fi.Name() != cas.dir {
			t.Errorf("want fi.Name()=%q; got %q (i=%d)", cas.dir, fi.Name(), i)
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
			mutfs := Must(TabTree(large))
			if err := mkdir(mutfs, cas, 0xD); err != nil {
				t.Errorf("want err=nil; got %q (i=%d, j=%d)", err, i, j)
				continue
			}
			if !Compare(Must(mutfs.Cd(cas)), Must(fs.Cd(cas))) {
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
		0:  {dir: filepath.FromSlash("/")},
		1:  {dir: filepath.FromSlash("/abc")},
		2:  {dir: filepath.FromSlash("/abc/1/2/3")},
		3:  {dir: filepath.FromSlash("/fs/abc")},
		4:  {dir: filepath.FromSlash("/fs/abc/1/2/3")},
		5:  {dir: filepath.FromSlash("/fs/memfs/abc/1/2/3")},
		6:  {dir: filepath.FromSlash("/fs/fs.go/testdata"), err: (*os.PathError)(nil)},
		7:  {dir: filepath.FromSlash("/LICENSE"), err: (*os.PathError)(nil)},
		8:  {dir: filepath.FromSlash("/README.md/testdata"), err: (*os.PathError)(nil)},
		9:  {dir: filepath.FromSlash("/fs/memfs/memfs.go/abc"), err: (*os.PathError)(nil)},
		10: {dir: filepath.FromSlash("/fs/memfs/memfs.go/abc/1/2/3"), err: (*os.PathError)(nil)},
	}
	for i, cas := range cases {
		err := fs.MkdirAll(cas.dir, 0xD)
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
		fi, err := fs.Stat(cas.dir)
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
		0: {path: filepath.FromSlash("c:/"), dir: true},
		1: {path: filepath.FromSlash("/"), dir: true},
		2: {path: filepath.FromSlash("/fs"), dir: true},
		3: {path: filepath.FromSlash("c:/fs/memfs"), dir: true},
		4: {path: filepath.FromSlash("/LICENSE"), dir: false},
		5: {path: filepath.FromSlash("c:/README.md"), dir: false},
		6: {path: filepath.FromSlash("/fs/fs.go"), dir: false},
		7: {path: filepath.FromSlash("c:/fs/memfs/memfs.go"), dir: false},
		8: {path: filepath.FromSlash("/fs/memfs/memfs_test.go"), dir: false},
	}
	for i, cas := range cases {
		f, err := fs.Open(cas.path)
		if err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		fi, err := f.Stat()
		if err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if fi.Name() != cas.path {
			t.Errorf("want fi.Name()=%q; got %q (i=%d)", cas.path, fi.Name(), i)
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
		0:  {file: filepath.FromSlash("/LICENSE")},
		1:  {file: filepath.FromSlash("/README.md")},
		2:  {file: filepath.FromSlash("/fs"), err: (*os.PathError)(nil)},
		3:  {file: filepath.FromSlash("/fs/fs.go")},
		4:  {file: filepath.FromSlash("/fs/memfs"), err: (*os.PathError)(nil)},
		5:  {file: filepath.FromSlash("/fs/memfs/memfs.go")},
		6:  {file: filepath.FromSlash("/fs/memfs/memfs_test.go")},
		7:  {file: filepath.FromSlash("/"), err: (*os.PathError)(nil)},
		8:  {file: filepath.FromSlash("c:"), err: (*os.PathError)(nil)},
		9:  {file: filepath.FromSlash("/er234"), err: os.ErrNotExist},
		10: {file: filepath.FromSlash("/fs/dfgdft345"), err: os.ErrNotExist},
	}
	for i, cas := range cases {
		err := fs.Remove(cas.file)
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
		if _, err := fs.Stat(cas.file); !os.IsNotExist(err) {
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
		filepath.FromSlash("/"): {
			{"fs", true},
			{"LICENSE", false},
			{"README.md", false},
		},
		filepath.FromSlash("/fs"): {
			{"fs.go", false},
			{"memfs", true},
		},
		filepath.FromSlash("c:/fs/memfs"): {
			{"memfs.go", false},
			{"memfs_test.go", false},
		},
	}
	for path, cas := range cases {
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
		rhs := Must(TabTree(cas.fs))
		lhs, err := fs.Cd(cas.path)
		if err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if !Compare(lhs, rhs) {
			t.Errorf("want Compare(...)=true; got false (i=%d)", i)
		}
	}
}
