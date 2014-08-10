package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/rjeczalik/tools/fs/memfs"
	"github.com/rjeczalik/tools/rw"
)

func nonnil(err ...error) error {
	for _, err := range err {
		if err != nil {
			return err
		}
	}
	return nil
}

var (
	begin = []byte(`memfs.Must(memfs.UnmarshalTab([]byte("`)
	sep   = []byte("\" +\n\t\"")
	end   = []byte("\")))\n")
)

var minwidth int = len(begin) + len(sep) + len(end) + 1

// ErrInvalidWidth signals requested column width is to small to represent
// memfs.FS as Go literal.
var ErrInvalidWidth = fmt.Errorf("gotree: invalid width size, minimum is %d", minwidth)

// EncodeLiteral represents fs as Go literal, encoding it to the w writer.
// The function tries to keep the literal max n characters wide, splitting it
// into multiple lines when necessary.
//
// Example
//
// For the following fs:
//
//   var fs = memfs.FS{
//              Tree: memfs.Directory{
//                "dir": memfs.Directory{
//                  "file.txt": memfs.File{},
//                },
//              },
//            }
//
//   EncodeLiteral(fs, 80, os.Stdout)
//
// The EncodeLiteral prints to standard output:
//
//   memfs.Must(memfs.UnmarshalTab([]byte(".\n\tdir\n\t\tfile.txt\n")))
func EncodeLiteral(fs memfs.FS, n int, w io.Writer) (err error) {
	if n < minwidth {
		err = ErrInvalidWidth
		return
	}
	pr, pw := io.Pipe()
	ch := make(chan error, 1)
	go func() {
		ch <- nonnil(memfs.Tab.Encode(fs, pw), pw.Close())

	}()
	scr := &rw.NinjaReader{
		N: n,
		R: io.MultiReader(
			bytes.NewReader(begin),
			&rw.QuotedReader{R: pr},
			bytes.NewReader(end),
		),
		Sep: sep,
	}
	_, err = io.Copy(w, scr)
	if e := <-ch; e != nil && err == nil {
		err = e
	}
	return
}
