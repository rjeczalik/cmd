package rw

import (
	"strings"
	"testing"
)

func TestEqual(t *testing.T) {
	cases := [...]struct {
		lhs, rhs string
		buflen   int
		ok       bool
	}{
		0: {
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			52,
			true,
		},
		1: {
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			2,
			true,
		},
		2: {
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			"ABCDEFGHIJ_LMNOPQRSTUVWXYZ",
			52,
			false,
		},
		3: {
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			"ABCDEFGHIJKLMNOPQRSTUVW_YZ",
			2,
			false,
		},
	}
	for i, cas := range cases {
		lhs := strings.NewReader(cas.lhs)
		rhs := strings.NewReader(cas.rhs)
		if ok := equal(lhs, rhs, cas.buflen); ok != cas.ok {
			t.Errorf("want equal(...)=%v; got %v (i=%d)", cas.ok, ok, i)
		}
	}
}
