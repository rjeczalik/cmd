package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/rjeczalik/tools/fs/memfs"
)

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

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

// QuotedReader proxies reads of R, quoting every string with strconv.Quote.
//
// Example
//
// The following line:
//
//   io.Copy(os.Stdout, &QuotedReader{R: strings.NewReader("\none\nline\n")})
//
// Prints to standard output:
//
//   \none\n\line\n
type QuotedReader struct {
	buf bytes.Buffer // buffers quoted bytes
	err error        // underlying reader error
	R   io.Reader    // underlying reader
}

// Read implements io.Reader.
func (qr *QuotedReader) Read(p []byte) (n int, err error) {
	if qr.err == nil {
		n, qr.err = qr.R.Read(p)
	}
	if n != 0 {
		s := strconv.Quote(string(p[:n]))
		qr.buf.WriteString(strings.Trim(s, "\""))
	}
	if n, err = qr.buf.Read(p); err == io.EOF {
		err = qr.err
	}
	return
}

// NinjaReader proxies reads of R, inserting Sep bytes to the stream each time
// N bytes was read from the underlying reader.
//
// It ensures Sep does not split escape sequences.
//
// TODO(rjeczalik): Current implementation does not play well with Go escape
//                  sequences longer than two characters (that is other than
//                  '\t', '\r', '\n' etc.).
type NinjaReader struct {
	N    int       // max bytes length read between separators
	R    io.Reader // underlying reader
	Sep  []byte    // separator
	n    int       // counter
	rsep io.Reader // sep reader
}

// Read implements io.Reader.
func (scr *NinjaReader) Read(p []byte) (n int, err error) {
	var (
		m   int // size of underlying reads
		l   int // bytes left to next separator
		lim = scr.N - len(scr.Sep) - 1
	)
	for n != len(p) {
		if scr.rsep != nil {
			m, err = scr.rsep.Read(p[n:])
			n, scr.n = n+m, scr.n+m
			if err != nil && err != io.EOF {
				break
			}
			if err == io.EOF {
				scr.rsep = nil
			}
		} else {
			l = min(lim-scr.n, len(p)-n)
			m, err = scr.R.Read(p[n : n+l])
			n, scr.n = n+m, scr.n+m
			if err != nil {
				break
			}
			if p[n-1] == '\\' {
				scr.n--
			}
			if scr.n == lim {
				scr.n = 0
				if len(scr.Sep) != 0 {
					scr.rsep = bytes.NewReader(scr.Sep)
				}
			}
		}
	}
	return
}

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
	scr := &NinjaReader{
		N: n,
		R: io.MultiReader(
			bytes.NewReader(begin),
			&QuotedReader{R: pr},
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
