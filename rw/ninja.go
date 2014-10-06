package rw

import (
	"bytes"
	"io"
)

// NinjedReader proxies reads of R, inserting Sep bytes to the stream each time
// N bytes was read from the underlying reader.
//
// It ensures Sep does not split escape sequences.
//
// TODO(rjeczalik): Current implementation does not play well with Go escape
// sequences longer than two characters (that is other than '\t', '\r', '\n' etc.).
type NinjedReader struct {
	R    io.Reader // underlying reader
	N    int       // max bytes length read between separators
	Sep  []byte    // separator
	n    int       // counter
	rsep io.Reader // sep reader
}

// NinjaReader creates a reader that interleaves reads from the r reader with sep
// after each n bytes read from the underlying reader.
func NinjaReader(r io.Reader, sep []byte, n int) io.Reader {
	return &NinjedReader{
		R:   r,
		N:   n,
		Sep: sep,
	}
}

// Read implements io.Reader.
func (nr *NinjedReader) Read(p []byte) (n int, err error) {
	var (
		m int // size of underlying reads
		l int // bytes left to next separator
	)
	for n != len(p) {
		if nr.rsep != nil {
			m, err = nr.rsep.Read(p[n:])
			n += m
			if err != nil && err != io.EOF {
				break
			}
			if err == io.EOF {
				nr.rsep = nil
			}
		} else {
			l = min(nr.N-nr.n, len(p)-n)
			m, err = nr.R.Read(p[n : n+l])
			n, nr.n = n+m, nr.n+m
			if m == 0 && err != nil {
				break
			}
			if p[n-1] == '\\' {
				nr.n--
			}
			if nr.n == nr.N {
				nr.n = 0
				if len(nr.Sep) != 0 {
					nr.rsep = bytes.NewReader(nr.Sep)
				}
			}
		}
	}
	return
}
