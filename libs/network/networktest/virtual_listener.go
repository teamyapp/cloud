package networktest

import (
	"net"
)

type VirtualListener struct {
	localAddress net.Addr
	connections  chan net.Conn
}

var _ net.Listener = (*VirtualListener)(nil)

func (v VirtualListener) Accept() (net.Conn, error) {
	return <-v.connections, nil
}

func (v VirtualListener) Close() error {
	close(v.connections)
	return nil
}

func (v VirtualListener) Addr() net.Addr {
	return v.localAddress
}

func (v VirtualListener) NotifyIncomingConnection(conn net.Conn) {
	go func() {
		v.connections <- conn
	}()
}

func newVirtualListener(localAddress net.Addr) VirtualListener {
	return VirtualListener{
		localAddress: localAddress,
		connections:  make(chan net.Conn, 0),
	}
}
