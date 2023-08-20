package networktest

import (
	"net"
)

type ProxyRoute struct {
	Endpoint    string
	MatchTarget func(addr net.Addr) bool
}
