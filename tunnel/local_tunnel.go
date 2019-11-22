package tunnel

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

type LocalTunnel struct {
	//代理目的地
	targetConn net.Conn
}

var LocalTunnelConf = struct {
	SocksForward string
}{}

func (t *LocalTunnel) Transport(src io.ReadWriter) error {
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

func NewLocalTunnel(dstHost string, dstPort string) (t Tunnel, err error) {
	var conn net.Conn
	if LocalTunnelConf.SocksForward == "" {
		conn, err = net.Dial("tcp", net.JoinHostPort(dstHost, dstPort))
	} else {
		log.Println("forward:", LocalTunnelConf.SocksForward)
		conn, err = socksForward(LocalTunnelConf.SocksForward, dstHost, dstPort)
	}
	if err != nil {
		return nil, err
	}
	return &LocalTunnel{conn}, nil
}
func (t *LocalTunnel) BindAddress() net.IP {
	return net.IPv4(127, 0, 0, 1)
}
func (t *LocalTunnel) BindPort() uint16 {
	n, err := strconv.Atoi(strings.Split(t.targetConn.LocalAddr().String(), ":")[1])
	if err != nil {
		return 0
	}
	return uint16(n)
}

//将请求转发给下一级代理
func socksForward(socksAddr, dstAddr, dstPort string) (socksConn net.Conn, err error) {
	conn, err := net.Dial("tcp", socksAddr)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if err = startSocksRequest(conn, dstAddr, dstPort); err != nil {
		return nil, err
	}
	return conn, nil
}

var methodSelectionMessage = []byte{0x05, 0x00}

//socks代理请求
func startSocksRequest(conn net.Conn, dstAddrStr, dstPortStr string) (err error) {
	// VER NMETHODS METHODS(NO_AUTH)
	_, err = conn.Write([]byte{0x05, 0x01, 0x00})
	if err != nil {
		return err
	}
	buf := [2]byte{}
	_, err = io.ReadFull(conn, buf[:])
	if err != nil {
		return err
	}
	if !bytes.Equal(methodSelectionMessage, buf[:]) {
		return fmt.Errorf("socks method is not acceptable")
	}

	request, err := makeRequestBytes(dstAddrStr, dstPortStr)
	if err != nil {
		return err
	}
	conn.Write(request)
	reply := make([]byte, 256)
	_, err = io.ReadFull(conn, reply[:5])
	if reply[2] != 0x0 {
		return fmt.Errorf("socks server refused the request")
	}
	last := 0
	switch reply[3] {
	case 0x1: //ipv4
		last = 5
	case 0x3: //domain
		last = int(reply[4]) + 1
	case 0x4: //ipv6
		last = 17
	}
	_, err = io.ReadFull(conn, reply[:last])
	return err
}

/*
   The SOCKS request is formed as follows:

        +----+-----+-------+------+----------+----------+
        |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
        +----+-----+-------+------+----------+----------+
        | 1  |  1  | X'00' |  1   | Variable |    2     |
        +----+-----+-------+------+----------+----------+
*/
func makeRequestBytes(dstAddrStr, dstPortStr string) ([]byte, error) {
	atype, dstAddr := convertDstAddr(dstAddrStr)
	dstPort, err := convertPort(dstPortStr)
	if err != nil {
		return nil, fmt.Errorf("convert DST.PORT error: %v", err)
	}
	requestLen := 6 + len(dstAddr)
	request := make([]byte, requestLen)
	request[0], request[1], request[2], request[3] = 0x05, 0x01, 0x0, atype
	copy(request[4:], dstAddr)
	request[requestLen-1], request[requestLen-2] = dstPort[1], dstPort[0]
	return request, nil
}

func convertDstAddr(addr string) (atype byte, dstAddr []byte) {
	ip := net.ParseIP(addr)
	if len(ip) == 0 {
		atype = 0x03 //o  DOMAINNAME: X'03'
		dstAddr := make([]byte, len(addr)+1)
		dstAddr[0] = byte(len(addr))
		copy(dstAddr[1:], addr)
		return atype, dstAddr
	}
	if ip4 := ip.To4(); len(ip4) == 4 {
		atype = 0x01 // o  IP V4 address: X'01'
		return atype, ip4
	}
	atype = 0x04 // o  IP V6 address: X'04'

	return atype, ip

}
func convertPort(portStr string) ([]byte, error) {
	buf := [2]byte{}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}
	buf[0], buf[1] = byte(port>>8), byte(port&0xff)
	return buf[:], err
}
