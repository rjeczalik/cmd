package fsutil

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/rjeczalik/tools/fs/memfs"
)

var tree = memfs.Must(memfs.UnmarshalTab([]byte(".\n\ta\n\t\tb1\n\t\t\tc1\n\t\t\t\tc" +
	"1.txt\n\t\t\tc2\n\t\t\t\tc2.txt\n\t\t\tc3\n\t\t\t\tc3.txt\n\t\t\t\t" +
	"d1\n\t\t\t\t\te1\n\t\t\t\t\t\t_\n\t\t\t\t\t\t\t_.txt\n\t\t\t\t\t\te" +
	"/\n\t\t\t\t\t\te1.txt\n\t\t\t\t\t\te2.txt\n\t\tb2\n\t\t\tc1\n\t\t\t" +
	"\td1.txt\n\t\t\t\td2/\n\t\t\t\td3.txt\n\ta.txt\n\tw\n\t\tw.txt\n\t\t" +
	"x\n\t\t\ty\n\t\t\t\tz\n\t\t\t\t\t1.txt\n\t\t\ty.txt\n")))

func TestTeeCreate(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestTeeMkdir(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestTeeMkdirAll(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

// TODO(rjeczalik): Split into TestTeeOpen and TestTeeReaddir
func TestTeeOpen(t *testing.T) {
	fmt.Println(tree)
	cases := [...]struct {
		open []string
		read []string
		fs   []byte
	}{
		0: {
			open: []string{"/w/x/y/z"},
			fs:   []byte(".\nw\n\tx\n\t\ty\n\t\t\tz/"),
		},
		1: {
			open: []string{"/a.txt", "/w/w.txt", "/a"},
			fs:   []byte(".\na/\na.txt\nw\n\tw.txt"),
		},
		2: {
			read: []string{"/a/b2/c1"},
			fs:   []byte(".\na\n\tb2\n\t\tc1\n\t\t\td1.txt\n\t\t\td2/\n\t\t\td3.txt"),
		},
		3: {
			read: []string{"/a/b1/c1", "/a/b1/c2", "/a/b1/c3"},
			fs: []byte(".\na\n\tb1\n\t\tc1\n\t\t\tc1.txt\n\t\tc2\n\t\t\tc2.txt\n\t\t" +
				"c3\n\t\t\tc3.txt\n\t\t\td1/"),
		},
		4: {
			read: []string{"/w", "/w/x/y", "/w/x/y/z", "/w/x"},
			fs:   []byte(".\nw\n\tw.txt\n\tx\n\t\ty\n\t\t\tz\n\t\t\t\t1.txt\n\t\ty.txt"),
		}}
LOOP:
	for i, cas := range cases {
		spy := memfs.New()
		tee := TeeFilesystem(tree, spy)
		for j, path := range cas.open {
			if _, err := tee.Open(filepath.FromSlash(path)); err != nil {
				t.Errorf("want err=nil; got %q (i=%d, j=%d)", err, i, j)
				continue LOOP
			}
		}
		for j, path := range cas.read {
			f, err := tee.Open(filepath.FromSlash(path))
			if err != nil {
				t.Errorf("want err=nil; got %q (i=%d, j=%d)", err, i, j)
				continue LOOP
			}
			if _, err = f.Readdir(0); err != nil {
				t.Errorf("want err=nil; got %q (i=%d, j=%d)", err, i, j)
				continue LOOP
			}
		}
		if !memfs.Equal(spy, memfs.Must(memfs.UnmarshalTab(cas.fs))) {
			t.Errorf("want Compare(...)=true; got false (i=%d)", i)
		}
	}
}

func TestTeeRemove(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestTeeRemoveAll(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestTeeStat(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestTeeWalk(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}
