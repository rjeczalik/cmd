package rw

import (
	"bytes"
	"io"
)

// Equal reads simultanously from lhs and rhs and compares bytes read. It returns
// bool if bytes read from both reads are equal.
func Equal(lhs, rhs io.Reader) bool {
	// Default buffer length, as used in io.Copy implementation.
	return equal(lhs, rhs, 32*1024)
}

func equal(lhs, rhs io.Reader, buflen int) bool {
	var (
		err  error
		piv  int = buflen / 2
		buf      = make([]byte, piv*2)
		n, m int
	)
	for {
		if n, err = lhs.Read(buf[:piv]); err != nil && err != io.EOF {
			return false
		}
		if m, err = rhs.Read(buf[piv:]); err != nil && err != io.EOF {
			return false
		}
		if n != m || !bytes.Equal(buf[:n], buf[piv:piv+m]) {
			return false
		}
		if err == io.EOF {
			break
		}
	}
	return true
}
