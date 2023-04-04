package networktest

import (
	"fmt"
	"net"
	"sync"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/network"
)

type Endpoint struct {
	transportProtocol string
	hostAndPort       string
}

var _ net.Addr = (*Endpoint)(nil)

func (e Endpoint) Network() string {
	return e.transportProtocol
}

func (e Endpoint) String() string {
	return e.hostAndPort
}

type VirtualNetwork struct {
	listenersMut        sync.RWMutex
	endpointToListeners map[string]VirtualListener
}

var _ network.Network = (*VirtualNetwork)(nil)

func (v *VirtualNetwork) Dial(transportProtocol string, hostAndPort string) (net.Conn, error) {
	v.listenersMut.RLock()
	defer v.listenersMut.RUnlock()
	listener, ok := v.endpointToListeners[hostAndPort]
	if !ok {
		return nil, fmt.Errorf("no is listening at endpoint: hostAndPort=%v", hostAndPort)
	}

	clientConn, serverConn := net.Pipe()
	listener.NotifyIncomingConnection(serverConn)
	return clientConn, nil
}

func (v *VirtualNetwork) Listen(transportProtocol string, hostAndPort string) (net.Listener, error) {
	endpoint := Endpoint{
		transportProtocol: transportProtocol,
		hostAndPort:       hostAndPort,
	}

	listener := newVirtualListener(endpoint)
	return listener, nil
}

func (v *VirtualNetwork) BindProxyEndpoint(hostAndPort string, listener net.Listener) *errs.Error {
	v.listenersMut.Lock()
	defer v.listenersMut.Unlock()
	_, ok := v.endpointToListeners[hostAndPort]
	if ok {
		return errs.NewError(errs.Unknown, fmt.Sprintf("endpoint already occupied: hostAndPort=%v", hostAndPort))
	}

	virtualListener, ok := listener.(VirtualListener)
	if !ok {
		return errs.NewError(errs.Unknown, "listener must be VirtualListener")
	}

	v.endpointToListeners[hostAndPort] = virtualListener
	return nil
}

func NewVirtualNetwork() *VirtualNetwork {
	return &VirtualNetwork{
		endpointToListeners: map[string]VirtualListener{},
	}
}
