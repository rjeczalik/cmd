package memnetz

import (
	"errors"
	"net"
	"sync"

	"github.com/rjeczalik/tools/netz"
)

var (
	errClosing = errors.New("netutil: use of closed network connection")
	errRefused = errors.New("netutil: connection refused")
	errUsing   = errors.New("netutil: address already in use")
	errNetwork = errors.New("netutil: invalid network")
)

type lis struct {
	addr net.Addr
	conn chan net.Conn
}

// TODO(rjeczalik): do not hardcode addr
func newLis(port uint16) *lis {
	return &lis{
		addr: &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: int(port)},
		conn: make(chan net.Conn, 1),
	}
}

func (l *lis) Accept() (net.Conn, error) {
	conn, ok := <-l.conn
	if !ok {
		return nil, errClosing
	}
	return conn, nil
}

func (l *lis) Close() (err error) {
	close(l.conn)
	return
}

func (l *lis) Addr() net.Addr {
	return l.addr
}

// TODO(rjeczalik): refactor to lis map getter instead
func netidx(network string) (idx uint8, err error) {
	switch network {
	case "tcp", "tcp4":
		idx = 0
	case "tcp6":
		idx = 1
	case "udp", "udp4":
		idx = 2
	case "udp6":
		idx = 3
	default:
		err = errNetwork
	}
	return
}

const maxUint16 uint16 = 1<<16 - 1

type network struct {
	mu   sync.RWMutex
	nets [4]map[uint16]*lis
}

func (n *network) portNum(network uint8, address string) (port uint16, err error) {
	if _, port, err = netz.SplitHostPort(address); err != nil {
		return
	}
	if port == 0 {
		n.mu.Lock()
		for {
			port = (port + 1) % maxUint16
			if port == 0 {
				port += 1
			}
			if _, ok := n.nets[network][port]; !ok {
				break
			}
		}
		n.mu.Unlock()
		return
	}
	return
}

func (n *network) Dial(network, addr string) (net.Conn, error) {
	num, err := netidx(network)
	if err != nil {
		return nil, err
	}
	port, err := n.portNum(num, addr)
	if err != nil {
		return nil, err
	}
	n.mu.RLock()
	l, ok := n.nets[num][port]
	n.mu.RUnlock()
	if !ok {
		return nil, errRefused
	}
	r, w := net.Pipe()
	l.conn <- r
	return w, nil
}

func (n *network) Listen(network, addr string) (net.Listener, error) {
	num, err := netidx(network)
	if err != nil {
		return nil, err
	}
	port, err := n.portNum(num, addr)
	if err != nil {
		return nil, err
	}
	n.mu.RLock()
	_, ok := n.nets[num][port]
	n.mu.RUnlock()
	if ok {
		return nil, errUsing
	}
	l := newLis(port)
	n.mu.Lock()
	n.nets[num][port] = l
	n.mu.Unlock()
	return l, nil
}

// NewNet provides an fake implementation for network.Network interface, which
// operates on a in-memory network.
func NewNet() netz.Network {
	n := &network{}
	for i := range n.nets {
		n.nets[i] = make(map[uint16]*lis)
	}
	return n
}

// Default is a default implementation of an in-memory fake for network.Network
// interface.
var Default netz.Network = NewNet()
