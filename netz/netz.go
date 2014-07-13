package netz

import (
	"net"
)

// Network provides an interface for the net package. It does not aim to be
// a complete interface, only those functions which are essential for mocking
// the network access.
type Network interface {
	// Dial connects to the address on the named network.
	Dial(net, addr string) (net.Conn, error)
	// Listen announces address on the local network.
	Listen(net, addr string) (net.Listener, error)
}

// Net provides an implementation for Network interface, wrapping functions
// from the net package.
type Net struct{}

var _ Network = Net{}

// Dial wraps net.Dial.
func (Net) Dial(network, addr string) (net.Conn, error) {
	return net.Dial(network, addr)
}

// Listen wraps net.Listen.
func (Net) Listen(network, addr string) (net.Listener, error) {
	return net.Listen(network, addr)
}

// Default is the default implementation of Network, which wraps functions
// the the net package.
var Default Network = Net{}
