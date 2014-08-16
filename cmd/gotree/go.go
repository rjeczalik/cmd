package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/rjeczalik/tools/fs/memfs"
	"github.com/rjeczalik/tools/rw"
)

var (
	begin = []byte(`memfs.Must(memfs.UnmarshalTab([]byte("`)
	sep   = []byte("\" +\n\t\"")
	end   = []byte("\")))\n")

	beginvar = `var %s = memfs.Must(memfs.UnmarshalTab([]byte("`
)

func minwidth(beg []byte) int {
	return len(beg) + len(sep) + len(end) + 1
}

func nonnil(err ...error) error {
	for _, err := range err {
		if err != nil {
			return err
		}
	}
	return nil
}

// InvalidWidthError signals requested column width is to small to represent
// memfs.FS as Go literal.
type InvalidWidthError struct {
	Min int
}

// Error implements errro interface.
func (iwe InvalidWidthError) Error() string {
	return "gotree: invalid width size, minimum is " + strconv.Itoa(iwe.Min)
}

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
//   EncodeLiteral(fs, 80, "", os.Stdout)
//
// The EncodeLiteral prints to standard output:
//
//   memfs.Must(memfs.UnmarshalTab([]byte(".\n\tdir\n\t\tfile.txt\n")))
//
// If v == "" then the output is Go literal, not assigned to any variable.
//
// If v != "" and n == 0 then the output is Go variable with column width set
// to 80.
//
// If v != "" and n != 0 then the output is Go variable printed with specified
// column width.
func EncodeLiteral(fs memfs.FS, n int, v string, w io.Writer) (err error) {
	var beg []byte = begin
	if v != "" {
		beg = []byte(fmt.Sprintf(beginvar, v))
	}
	if n == 0 {
		n = 80
	}
	if min := minwidth(beg); n < min {
		return InvalidWidthError{Min: min}
	}
	pr, pw := io.Pipe()
	ch := make(chan error, 1)
	go func() {
		ch <- nonnil(memfs.Tab.Encode(fs, pw), pw.Close())
	}()
	nr := rw.NinjaReader(
		io.MultiReader(
			bytes.NewReader(beg),
			rw.QuoteReader(pr),
			bytes.NewReader(end),
		),
		sep, n,
	)
	_, err = io.Copy(w, nr)
	if e := <-ch; e != nil && err == nil {
		err = e
	}
	return
}
