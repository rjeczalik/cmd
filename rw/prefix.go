package rw

import (
	"bytes"
	"io"
	"strings"
)

// PrefixedWriter TODO
type PrefixedWriter struct {
	W      io.Writer     // underlying writer
	Prefix func() string // generates string used to prefix each line

	first bool
}

// PrefixWriter TODO
func PrefixWriter(writer io.Writer, prefix func() string) io.Writer {
	return &PrefixedWriter{W: writer, Prefix: prefix}
}

// Write TODO
func (pw *PrefixedWriter) Write(p []byte) (int, error) {
	var prefix string
	if !pw.first {
		prefix = pw.Prefix()
		// Write prefix for the first line.
		if _, err := io.Copy(pw.W, strings.NewReader(prefix)); err != nil {
			return 0, err
		}
		pw.first = true
	}
	var i, j, k, n int
	if j, n = indexnl(p); i != -1 {
		var (
			err error
			buf bytes.Buffer
		)
		if prefix == "" {
			prefix = pw.Prefix()
		}
		// If p contains multiple newlines we loop over each of them.
	Loop:
		// Write line.
		if _, err = buf.Write(p[i : j+1]); err != nil {
			return 0, err
		}
		// Write prefix for the next line.
		if _, err = buf.WriteString(prefix); err != nil {
			return 0, err
		}
		// Search next newline.
		if k, n = indexnl(p[j+n+1:]); k != -1 {
			i, j = j+1, j+k+n+1
			goto Loop
		}
		// Write last line.
		if _, err = buf.Write(p[j+n+1:]); err != nil {
			return 0, err
		}
		n, err := io.Copy(pw.W, &buf)
		return min(int(n), len(p)), err
	}
	return pw.W.Write(p)
}
