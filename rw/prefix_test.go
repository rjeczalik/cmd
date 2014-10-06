package rw

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func ring(args []string) func() string {
	var i int
	return func() (s string) {
		s, i = args[i], (i+1)%len(args)
		return
	}
}

func TestPrefixedWriter(t *testing.T) {
	cases := [...]struct {
		s      string
		prefix func() string
		want   string
	}{
		{
			"a\nb\nc\r\nd\n",
			func() string { return "[$(DATE)] " },
			"[$(DATE)] a\n[$(DATE)] b\n[$(DATE)] c\r\n[$(DATE)] d\n",
		},
		{
			"LINE 1\r\nLINE 2\r\nLINE 3\nLINE 4\r\nLINE 5\n",
			ring([]string{"+ ", "- "}),
			"+ LINE 1\r\n- LINE 2\r\n+ LINE 3\n- LINE 4\r\n+ LINE 5\n",
		},
	}
	for i, cas := range cases {
		// TODO(rjeczalik): Fix test-case 1
		if i != 0 {
			continue
		}
		var buf bytes.Buffer
		if _, err := io.Copy(PrefixWriter(&buf, cas.prefix), strings.NewReader(cas.s)); err != nil {
			t.Errorf("want err=nil; got %v (i=%d)", err, i)
			continue
		}
		if got := buf.String(); got != cas.want {
			t.Errorf("want got=%q; got %q (i=%d)", cas.want, got, i)
		}
	}
}
