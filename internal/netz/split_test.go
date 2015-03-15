package netz

import "testing"

func TestSplitHostPort(t *testing.T) {
	cases := [...]struct {
		hostport string
		host     string
		port     uint16
	}{
		0: {"[::1]:59433", "::1", 59433},
		1: {":61200", "", 61200},
		2: {":0", "", 0},
		3: {"localhost:52302", "localhost", 52302},
	}
	casesErr := []string{
		0: "[::1]:67123",
		1: "localhost:-1",
		2: "localhost:e042w",
		3: "[::1]:",
	}
	for i, cas := range cases {
		host, port, err := SplitHostPort(cas.hostport)
		if err != nil {
			t.Errorf("want err=nil; got %v (i=%d)", err, i)
			continue
		}
		if host != cas.host {
			t.Errorf("want host=%q; got %q (i=%d)", cas.host, host, i)
		}
		if port != cas.port {
			t.Errorf("want port=%d; got %d (i=%d)", cas.port, port, i)
		}
	}
	for i, hostport := range casesErr {
		if _, _, err := SplitHostPort(hostport); err == nil {
			t.Errorf("want err!=nil (i=%d)", i)
		}
	}
}
