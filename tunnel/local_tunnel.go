package tunnel

import (
	"io"
	"net"
	"strconv"
	"strings"
)

type LocalTransporter struct {
	//代理目的地
	targetConn net.Conn
}

func (t *LocalTransporter) Transport(src io.ReadWriter) error {
	defer t.targetConn.Close()
	return doProxy(src, t.targetConn)
}

func doProxy(src io.ReadWriter, dst io.ReadWriter) error {
	ch := make(chan error, 2)
	go func() {
		ch <- proxy(dst, src)
	}()
	go func() {
		ch <- proxy(src, dst)
	}()
	for i := 0; i < 2; i++ {
		err := <-ch
		if err != nil {
			return err
		}
	}
	return nil

}
func proxy(src io.Reader, dst io.Writer) error {
	_, err := io.Copy(dst, src)
	return err
}

func NewLocalTransport(dstHost string, dstPort string) (Transporter, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(dstHost, dstPort))
	if err != nil {
		return nil, err
	}
	return &LocalTransporter{conn}, nil
}

func (t *LocalTransporter) BindAddress() net.IP {
	return net.IPv4(127, 0, 0, 1)
}
func (t *LocalTransporter) BindPort() uint16 {
	n, err := strconv.Atoi(strings.Split(t.targetConn.LocalAddr().String(), ":")[1])
	if err != nil {
		return 0
	}
	return uint16(n)
}
