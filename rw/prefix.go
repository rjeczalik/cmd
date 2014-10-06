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
	if !pw.first {
		// Write prefix for the first line.
		if _, err := io.Copy(pw.W, strings.NewReader(pw.Prefix())); err != nil {
			return 0, err
		}
		pw.first = true
	}
	var (
		i int // offset of the begining of a line
		j int // offset of the end of a line before newline characters
		k int // relative offset of a next begining of a line
		n int // width of the newline characters (1 for LF, 2 for CRLF)
	)
	if j, n = indexnl(p); i != -1 {
		var (
			err error
			buf bytes.Buffer
		)
		// If p contains multiple newlines we loop over each of them.
	Loop:
		// Write line.
		if _, err = buf.Write(p[i : j+n]); err != nil {
			return 0, err
		}
		i = j + n
		// Search next newline.
		if k, n = indexnl(p[i:]); k != -1 {
			// Write prefix for the next line.
			if _, err = buf.WriteString(pw.Prefix()); err != nil {
				return 0, err
			}
			j = i + k
			goto Loop
		}
		// Write last line if p does not end with a newline.
		if i < len(p) {
			// Write prefix for the last line.
			if _, err = buf.WriteString(pw.Prefix()); err != nil {
				return 0, err
			}
			// Write the last line.
			if _, err = buf.Write(p[i:]); err != nil {
				return 0, err
			}
		}
		n, err := io.Copy(pw.W, &buf)
		return min(int(n), len(p)), err
	}
	return pw.W.Write(p)
}
