package tunnel

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
)

type commonServer struct {
	Key string
}

func (c *commonServer) handle(rw io.ReadWriteCloser) (err error) {
	defer rw.Close()
	erw, err := NewEncodeReadWriter(rw, c.Key)
	if err != nil {
		return err
	}
	header := [6]byte{}
	_, err = io.ReadFull(erw, header[:])
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(header[:], headerSign) {
		return fmt.Errorf("wrong proxy header:%s", header)
	}
	addrLen := int(header[5])
	portAddr := make([]byte, addrLen)
	_, err = io.ReadFull(erw, portAddr)
	if err != nil {
		return err
	}
	port := strconv.Itoa(int(portAddr[0])<<8 + int(portAddr[1]))
	addr := string(portAddr[2:])

	local, err := NewLocalTunnel(addr, port)
	if err != nil {
		return err
	}
	return local.Transport(erw)

}

type commonClient struct {
	Key     string
	dstPort string
	dstAddr string
}

func (c *commonClient) BindAddress() net.IP {
	return net.IPv4(0, 0, 0, 0)
}
func (c *commonClient) BindPort() uint16 {
	return 0
}

func (c *commonClient) doTransport(server io.ReadWriter, src io.ReadWriter) (err error) {
	erw, err := NewEncodeReadWriter(server, c.Key)
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(c.dstPort)
	if err != nil {
		return fmt.Errorf("wrong port ,%v", err)
	}
	portBytes := [2]byte{byte(port >> 8), byte(port & 0xff)}
	header := make([]byte, 8+len(c.dstAddr))
	copy(header, headerSign)
	header[5] = byte(len(header) - 6)
	copy(header[6:], portBytes[:])
	copy(header[8:], c.dstAddr)
	erw.Write(header)
	return doProxy(src, erw)
}
