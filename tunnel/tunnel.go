package tunnel

import (
	"io"
	"net"
)

type Transporter interface {
	Transport(writer io.ReadWriter) error
	BindAddress() net.IP
	BindPort() uint16
}

func NewTransConf(form, key, host, port string) *TransConf {
	t := new(TransConf)
	t.form = form
	t.key = key
	t.serverPort = port
	t.serverHost = host
	return t
}

type TransConf struct {
	form       string //  local remote
	key        string
	serverHost string
	serverPort string
}

func (c *TransConf) NewTransporter(dstAddr string, dstPort string) (Transporter, error) {
	if c == nil {
		return NewLocalTransport(dstAddr, dstPort)
	}
	switch c.form {
	case "local":
		return NewLocalTransport(dstAddr, dstPort)
	default:
		//remote
		return NewRemoteTransporter(c.key, c.serverHost, c.serverPort, dstAddr, dstPort)
	}
}
