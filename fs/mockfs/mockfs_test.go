package mockfs

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func fixture() FS {
	return FS{
		Tree: Directory{
			"fs": Directory{
				"fs.go": NewFile("fs.go"),
				"mockfs": Directory{
					"mockfs.go":      NewFile("mockfs.go"),
					"mockfs_test.go": NewFile("mockfs_test.go"),
				},
			},
			"LICENSE":   NewFile("LICENSE"),
			"README.md": NewFile("README.md"),
		},
	}
}

func path(s string) string {
	if os.PathSeparator != '/' {
		s = strings.Replace(s, "/", "\\", -1)
	}
	return s
}

func TestCreate(t *testing.T) {
	fs := fixture()
	cases := [...]struct {
		file string
		err  error
	}{
		0:  {file: path("c:/fs/mockfs/all_test.go")},
		1:  {file: path("/LICENSE")},
		2:  {file: path("c:/fs/fs.go")},
		3:  {file: path("/LICENSE.md")},
		4:  {file: path("/fs/fs_test.go")},
		5:  {file: path("/"), err: (*os.PathError)(nil)},
		6:  {file: path("c:"), err: (*os.PathError)(nil)},
		7:  {file: path("c:/"), err: (*os.PathError)(nil)},
		8:  {file: path("/fs"), err: (*os.PathError)(nil)},
		9:  {file: path("/fs/mockfs"), err: (*os.PathError)(nil)},
		10: {file: path("/.git/config"), err: (*os.PathError)(nil)},
		11: {file: path("/fs/.svn/config"), err: (*os.PathError)(nil)},
		12: {file: path("/LICENSE/OTHER.md"), err: (*os.PathError)(nil)},
		13: {file: path("/fs/fs.go/detail.go"), err: (*os.PathError)(nil)},
		14: {file: path("/fs/mockfs/nfs/nfs.go"), err: (*os.PathError)(nil)},
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
	fs := fixture()
	cases := [...]struct {
		dir string
		err error
	}{
		0:  {dir: path("/testdata")},
		1:  {dir: path("/fs/testdata")},
		2:  {dir: path("c:/fs/mockfs/testdata")},
		3:  {dir: path("c:/testdata")},
		4:  {dir: path("c:/")},
		5:  {dir: path("/")},
		6:  {dir: path("c:/LICENSE"), err: (*os.PathError)(nil)},
		7:  {dir: path("c:/LICENSE/testdata"), err: (*os.PathError)(nil)},
		8:  {dir: path("/fs/mockfs/mockfs.go"), err: (*os.PathError)(nil)},
		9:  {dir: path("/fs/fs.go/testdata"), err: (*os.PathError)(nil)},
		10: {dir: path("c:/fs/mockfs/mockfs_test.go"), err: (*os.PathError)(nil)},
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

func TestOpen(t *testing.T) {
	fs := fixture()
	cases := [...]struct {
		path string
		dir  bool
	}{
		0: {path: path("c:/"), dir: true},
		1: {path: path("/"), dir: true},
		2: {path: path("/fs"), dir: true},
		3: {path: path("c:/fs/mockfs"), dir: true},
		4: {path: path("/LICENSE"), dir: false},
		5: {path: path("c:/README.md"), dir: false},
		6: {path: path("/fs/fs.go"), dir: false},
		7: {path: path("c:/fs/mockfs/mockfs.go"), dir: false},
		8: {path: path("/fs/mockfs/mockfs_test.go"), dir: false},
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
	fs := fixture()
	cases := [...]struct {
		file string
		err  error
	}{
		0:  {file: path("/LICENSE")},
		1:  {file: path("/README.md")},
		2:  {file: path("/fs"), err: (*os.PathError)(nil)},
		3:  {file: path("/fs/fs.go")},
		4:  {file: path("/fs/mockfs"), err: (*os.PathError)(nil)},
		5:  {file: path("/fs/mockfs/mockfs.go")},
		6:  {file: path("/fs/mockfs/mockfs_test.go")},
		7:  {file: path("/"), err: (*os.PathError)(nil)},
		8:  {file: path("c:"), err: (*os.PathError)(nil)},
		9:  {file: path("/er234"), err: os.ErrNotExist},
		10: {file: path("/fs/dfgdft345"), err: os.ErrNotExist},
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
	fs := fixture()
	cases := map[string][]struct {
		name string
		dir  bool
	}{
		path("/"): {
			{"fs", true},
			{"LICENSE", false},
			{"README.md", false},
		},
		path("/fs"): {
			{"fs.go", false},
			{"mockfs", true},
		},
		path("c:/fs/mockfs"): {
			{"mockfs.go", false},
			{"mockfs_test.go", false},
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
