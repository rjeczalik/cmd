package rw

import (
	"io"
	"os"
	"strings"
	"testing"
)

func ExampleQuoteReader() {
	io.Copy(os.Stdout, QuoteReader(strings.NewReader("\none\nline\n")))
	// Output:
	// \none\nline\n
}

func TestQuoteReader(t *testing.T) {
	cases := [...]struct {
		s string
		o string
	}{
		{"\none\nline\n", "\\none\\nline\\n"},
		{"\n\r\t\n", "\\n\\r\\t\\n"},
		{"", ""},
		{"abcdefghijklmn\n\n", "abcdefghijklmn\\n\\n"},
	}
	for i, cas := range cases {
		lhs := QuoteReader(strings.NewReader(cas.s))
		rhs := strings.NewReader(cas.o)
		if !Equal(lhs, rhs) {
			t.Errorf("want Equal(...)=true; got false (i=%d)", i)
		}
	}
}
