package rw

import (
	"io"
	"os"
	"strings"
	"testing"
)

func ExampleNinjaReader() {
	io.Copy(os.Stdout, NinjaReader(strings.NewReader("123"), []byte{'\n'}, 1))
	// Output:
	// 1
	// 2
	// 3
}

func TestNinjaReader(t *testing.T) {
	cases := [...]struct {
		s   string
		sep []byte
		n   int
		o   string
	}{
		0: {"123", []byte{'\n'}, 1, "1\n2\n3\n"},
		1: {"ABCDEFGHIJKL", []byte{'\r', '\n'}, 4, "ABCD\r\nEFGH\r\nIJKL\r\n"},
		2: {"AAAA", []byte("BBCCBB"), 2, "AABBCCBBAABBCCBB"},
	}
	for i, cas := range cases {
		lhs := NinjaReader(strings.NewReader(cas.s), cas.sep, cas.n)
		rhs := strings.NewReader(cas.o)
		if !Equal(lhs, rhs) {
			t.Errorf("want Equal(...)=true; got false (i=%d)", i)
		}
	}
}
