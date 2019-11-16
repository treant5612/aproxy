package socks5

import (
	"aproxy/transport"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

type SocksServer struct {
	method    byte
	transConf *transport.TransConf
}

func (s *SocksServer) newTransporter(host string, port string) (transport.Transporter, error) {
	return s.transConf.NewTransporter(host, port)
}
func NewServer(method string, transConf *transport.TransConf) *SocksServer {
	s := new(SocksServer)
	switch method {
	case "PASSWORD":
		s.method = METHOD_USER_PWD
	default:
		s.method = METHOD_NO_AUTH
	}
	s.transConf = transConf
	return s

}

func (s *SocksServer) Method() byte {
	return s.method
}

const (
	SocksVer  byte = 0x05
	SocksVer4 byte = 0x04
)
const socks4_request_granted byte = 90

const (
	METHOD_NO_AUTH       byte = 0x00
	METHOD_GSSAPI        byte = 0x01
	METHOD_USER_PWD      byte = 0x03
	METHOD_NO_ACCEPTABLE byte = 0xff
)

type SocksRequest struct {
	*SocksServer
	source    net.Conn
	bufReader *bufio.Reader
	version   byte
	cmd       byte //
	rsv       byte
	atyp      byte
	DstAddr   string
	DstPort   string
	trans     transport.Transporter
	//for socks4
	addrBytes []byte
	portBytes []byte
}

const (
	CMD_CONNECT       byte = 0x01
	CMD_BIND          byte = 0x02
	CMD_UDP_ASSOCIATE byte = 0x03
)
const (
	ATYPE_IPV4   byte = 0x01
	ATYPE_DOMAIN byte = 0x03
	ATYPE_IPV6   byte = 0x04
)

type SocksError struct {
	describe string
	response []byte
}

func (s SocksError) Error() string {
	return s.describe
}

var methodNotSupported = SocksError{
	describe: "method not supported",
	response: []byte{SocksVer, METHOD_NO_ACCEPTABLE},
}

func (s *SocksServer) NewSocksRequest(conn net.Conn) *SocksRequest {
	return &SocksRequest{SocksServer: s, source: conn, bufReader: bufio.NewReader(conn)}
}
func (s *SocksServer) ListenAndServe(network string, address string) error {
	listener, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		log.Println("Accept from ", conn.RemoteAddr())
		go s.handle(conn)
	}
	return nil
}
func (s *SocksServer) handle(conn net.Conn) {
	defer conn.Close()
	request := s.NewSocksRequest(conn)
	err := s.handleSocks(request)
	if err != nil {
		if sErr, ok := err.(SocksError); ok {
			request.reply(sErr.response)
		}
	}
	log.Println(err)
}

//socks5 protocol :rfc1928
func (s *SocksServer) handleSocks(request *SocksRequest) error {
	//验证
	err := request.authentication()
	if err != nil {
		return err
	}
	//接收请求
	err = request.requestDetail()
	if err != nil {
		return err
	}
	log.Printf("proxy: %s : %s\n", request.DstAddr, request.DstPort)
	//准备代理数据传输
	err = request.BuildTransport()
	if err != nil {
		request.requestRespond(1)
		return err
	}
	err = request.requestRespond(0)
	if err != nil {
		return err
	}
	//开始代理数据传输
	err = request.trans.Transport(request.source)
	return err
}

func (r *SocksRequest) reply(data []byte) (int, error) {
	return r.source.Write(data)
}

//验证过程
func (r *SocksRequest) authentication() error {
	b := make([]byte, 64)
	n, err := r.bufReader.Read(b)
	if err != nil {
		return err
	}
	b = b[:n]
	r.version = b[0]
	log.Println(r.version)
	if r.version != SocksVer && r.version != SocksVer4 {
		return methodNotSupported
	}
	if r.version == SocksVer4 {
		socks4Request(b, r)
		return nil
	}

	if !bytes.ContainsRune(b[2:], rune(r.Method())) {
		return methodNotSupported
	}
	_, err = r.reply([]byte{r.version, r.Method()})
	if err != nil {
		return err
	}
	/*
		switch r.Method()
		TODO
			GSSAPI
			USERNAME/PASSWORD
	*/
	return nil
}

func socks4Request(b []byte, r *SocksRequest) {

	//+----+----+----+----+----+----+----+----+----+----+....+----+
	//| VN | CD | DSTPORT |      DSTIP        | USERID       |NULL|
	//	+----+----+----+----+----+----+----+----+----+----+....+----+
	//   1    1      2              4           variable       1
	r.method = METHOD_NO_AUTH
	r.cmd = b[1]
	r.DstPort = strconv.Itoa(int(b[2])<<8 | int(b[3]))
	r.DstAddr = fmt.Sprintf("%d.%d.%d.%d", b[4], b[5], b[6], b[7])
	r.addrBytes, r.portBytes = make([]byte, 4), make([]byte, 2)
	copy(r.addrBytes, b[4:8])
	copy(r.portBytes, b[2:4])
	return
}

func (r *SocksRequest) requestDetail() (err error) {
	if r.version == SocksVer4 {
		return nil
	}
	/*
		+-----+-----+-------+------+----------+----------+
		| VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
		+-----+-----+-------+------+----------+----------+
		|  1  |  1  | X'00' |  1   | Variable |    2     |
		+-----+-----+-------+------+----------+----------+
	*/
	buf := [4]byte{}
	_, err = io.ReadFull(r.bufReader, buf[:])
	if err != nil {
		return err
	}
	r.version, r.cmd, r.rsv, r.atyp = buf[0], buf[1], buf[2], buf[3]
	r.DstAddr, err = readAddr(r.bufReader, r.atyp)
	if err != nil {
		return err
	}
	port := [2]byte{}
	_, err = io.ReadFull(r.bufReader, port[:])
	if err != nil {
		return err
	}
	r.DstPort = strconv.Itoa(int(port[0])<<8 | int(port[1]))
	return nil

}
func readAddr(r *bufio.Reader, atype byte) (string, error) {
	var addrLen byte = 0
	switch atype {
	case ATYPE_IPV4:
		addrLen = 4
	case ATYPE_DOMAIN:
		b, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		addrLen = b
	case ATYPE_IPV6:
		addrLen = 16
	}
	buf := make([]byte, addrLen)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}
	if atype == ATYPE_DOMAIN {
		return string(buf), nil
	}
	return net.IP(buf).String(), nil
}

func (r *SocksRequest) BuildTransport() (err error) {
	trans, err := r.newTransporter(r.DstAddr, r.DstPort)
	if err != nil {
		return err
	}
	r.trans = trans
	return nil
}
func socks4Reply(r *SocksRequest) []byte {
	//+----+----+----+----+----+----+----+----+
	//| VN | CD | DSTPORT |      DSTIP        |
	//	+----+----+----+----+----+----+----+----+
	//  1    1      2              4
	reply := [8]byte{}
	reply[0] = r.version
	reply[1] = socks4_request_granted
	copy(reply[2:4], r.portBytes)
	copy(reply[4:8], r.addrBytes)
	return reply[:]
}
func (r *SocksRequest) requestRespond(rep int) error {
	if r.version == SocksVer4 {
		reply := socks4Reply(r)
		_, err := r.reply(reply)
		return err
	}

	//	+----+-----+-------+------+----------+----------+
	//	|VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
	//	+----+-----+-------+------+----------+----------+
	//	| 1  |  1  | X'00' |  1   | Variable |    2     |
	//	+----+-----+-------+------+----------+----------+
	response := []byte{0x05, 0x0, 0x0, ATYPE_IPV4, 0, 0, 0, 0, 0, 0}
	switch rep {
	case 0:
	default:
		response[1] = 0x01
		_, err := r.reply(response)
		return err
	}
	addr, port := r.trans.BindAddress(), r.trans.BindPort()
	copy(addr[0:4], response[4:8])
	response[len(response)-2], response[len(response)-1] = byte(port>>8), byte(port&0xff)
	_, err := r.reply(response)
	if err != nil {
		return err
	}
	return nil
}
