package tunnel

import (
	"aproxy/filter"
	"io"
	"net"
)

type Tunnel interface {
	Transport(writer io.ReadWriter) error
	BindAddress() net.IP
	BindPort() uint16
}

func NewTunnelConf(form, key, host, port string) *Conf {
	t := new(Conf)
	t.form = form
	t.key = key
	t.serverPort = port
	t.serverHost = host
	return t
}

type Conf struct {
	form       string //  local remote
	key        string
	serverHost string
	serverPort string
}

func (c *Conf) NewTunnel(dstAddr string, dstPort string) (Tunnel, error) {
	if c == nil {
		return NewLocalTunnel(dstAddr, dstPort)
	}
	if filter.Proxy(dstAddr) {
		return NewRemoteTunnel(c.key, c.serverHost, c.serverPort, dstAddr, dstPort)
	}
	return NewLocalTunnel(dstAddr, dstPort)
}
