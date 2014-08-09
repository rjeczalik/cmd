package fsutil

import (
	"path/filepath"
	"testing"

	"github.com/rjeczalik/tools/fs/memfs"
)

var tree = []byte(".\na\n\tb1\n\t\tc1\n\t\t\tc1.txt\n\t\tc2\n\t\t\tc2.txt\n\t\t" +
	"c3\n\t\t\tc3.txt\n\t\t\td1\n\t\t\t\te1\n\t\t\t\t\t_\n\t\t\t\t\t\t_.txt" +
	"\n\t\t\t\t\te1.txt\n\t\t\t\t\te2.txt\n\t\t\t\t\te/\n\tb2\n\t\tc1\n\t\t" +
	"\td1.txt\n\t\t\td2/\n\t\t\td3.txt\na.txt\nw\n\tw.txt\n\tx\n\t\ty\n\t\t" +
	"\tz\n\t\t\t\t1.txt\n\t\ty.txt\n")

func TestTee(t *testing.T) {
	fs := memfs.Must(memfs.UnmarshalTab(tree))
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
		if i != 4 {
			continue
		}
		spy := memfs.New()
		tee := TeeFilesystem(fs, spy)
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
