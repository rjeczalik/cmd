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
			"LINE 1\r\nLINE 2\r\nLINE 3\nLINE 4\r\n",
			ring([]string{"+ ", "- "}),
			"+ LINE 1\r\n- LINE 2\r\n+ LINE 3\n- LINE 4\r\n",
		},
		{
			"\n\n\r\n\r\n\n\r\n",
			ring([]string{"xxx", "XXX", "..."}),
			"xxx\nXXX\n...\r\nxxx\r\nXXX\n...\r\n",
		},
		{
			"asd qwe qwe 123 \r\n werwer fq34234 234 \n dfg dfg dfg dfg",
			func() string { return "9 123 012 30: " },
			"9 123 012 30: asd qwe qwe 123 \r\n9 123 012 30:  werwer fq34234 " +
				"234 \n9 123 012 30:  dfg dfg dfg dfg",
		},
	}
	for i, cas := range cases {
		var buf bytes.Buffer
		n, err := io.Copy(PrefixWriter(&buf, cas.prefix), strings.NewReader(cas.s))
		if err != nil {
			t.Errorf("want err=nil; got %v (i=%d)", err, i)
			continue
		}
		if got := buf.String(); got != cas.want {
			t.Errorf("want got=%q; got %q (i=%d)", cas.want, got, i)
			continue
		}
		if want := len(cas.s); int(n) != want {
			t.Errorf("want n=%d; got %d (i=%d)", want, n, i)
		}
	}
}
