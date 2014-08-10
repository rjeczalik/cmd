package rw

import (
	"bytes"
	"io"
)

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
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
