package rw

import (
	"bytes"
	"io"
	"strconv"
)

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

// QuoteReader creates a reader that escapes every control character read from
// the underlying reader with its corresponding escape sequence.
func QuoteReader(r io.Reader) io.Reader {
	return &QuotedReader{
		R: r,
	}
}

// Read implements io.Reader.
func (qr *QuotedReader) Read(p []byte) (n int, err error) {
	if qr.err == nil {
		n, qr.err = qr.R.Read(p)
	}
	if n != 0 {
		s := strconv.Quote(string(p[:n]))
		qr.buf.WriteString(s[1 : len(s)-1])
	}
	if n, err = qr.buf.Read(p); err == io.EOF {
		err = qr.err
	}
	return
}
