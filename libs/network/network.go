package network

import (
	"net"
)

type Network interface {
	Dial(transportProtocol string, hostAndPort string) (net.Conn, error)
	Listen(transportProtocol string, hostAndPort string) (net.Listener, error)
}

type Socket struct {
}

var _ Network = (*Socket)(nil)

func (s Socket) Dial(transportProtocol string, hostAndPort string) (net.Conn, error) {
	return net.Dial(transportProtocol, hostAndPort)
}

func (s Socket) Listen(transportProtocol string, ipAndPort string) (net.Listener, error) {
	return net.Listen(transportProtocol, ipAndPort)
}

func NewSocket() Socket {
	return Socket{}
}
