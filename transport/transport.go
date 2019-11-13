package transport

import (
	"io"
	"net"
)

type Transporter interface {
	Transport(writer io.ReadWriter) error
	BindAddress()net.IP
	BindPort() uint16
}
