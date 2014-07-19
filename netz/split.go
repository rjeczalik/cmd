package netz

import (
	"net"
	"strconv"
)

// SplitHostPort is a helper function for net.SplitHostPort, which returns
// a port as uint16 instead of a string.
func SplitHostPort(hostport string) (string, uint16, error) {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return "", 0, err
	}
	n, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return "", 0, err
	}
	return host, uint16(n), nil
}
