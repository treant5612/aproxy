package transport

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestEncodeReadWriter_Net(t *testing.T) {
	key := "#@!iselsk"
	listener, err := net.Listen("tcp", ":4887")
	if err != nil {
		t.FailNow()
	}
	go func() {
		conn, _ := net.Dial("tcp", "localhost:4887")
		erw0, _ := NewEncodeReadWriter(conn, key)
		defer conn.Close()
		for i := 128; i < 2048; i++ {
			erw0.Write([]byte(strings.Repeat("A", i)))
		}
	}()
	conn, err := listener.Accept()
	if err != nil {
		t.FailNow()
	}
	erw, err := NewEncodeReadWriter(conn, key)
	bufReader := bufio.NewReader(erw)
	buf := make([]byte, 256)
	for {
		n, err := bufReader.Read(buf)
		if err != nil {
			break
		}
		fmt.Printf("%q\n", buf[:n])
	}
}
func TestEncodeReadWriter(t *testing.T) {
	raw0 := []byte("你好TestEncodeReadWriter")
	raw1 := []byte("test str2")

	bytesrw := newBytesRw(nil)
	erw, err := NewEncodeReadWriter(bytesrw, "#@!iselsk")
	if err != nil {
		t.FailNow()
	}
	erw.Write(raw0)
	p := make([]byte, 256)
	n, _ := erw.Read(p)
	if !bytes.Equal(raw0, p[:n]) {
		fmt.Printf("raw0(%q) != plainText(%q)", raw0, p)
		t.Fail()
	}

	erw.Write(raw1)
	p = make([]byte, 256)
	n, _ = erw.Read(p)
	if !bytes.Equal(raw1, p[:n]) {
		fmt.Printf("raw0(%q) != plainText(%q)", raw1, p)
		t.Fail()
	}

}

func TestRemoteTransporterServer(t *testing.T) {
	server := &RemoteTransporterServer{Key: "testkey"}
	go server.ListenAndServe("tcp",":4721")


}

type bytesRw struct {
	data []byte
}

func (brw *bytesRw) Read(b []byte) (n int, err error) {
	return bytes.NewReader(brw.data).Read(b)
}
func newBytesRw(p []byte) *bytesRw {
	return &bytesRw{
		p,
	}
}
func (brw *bytesRw) Write(p []byte) (n int, err error) {
	brw.data = p
	return len(brw.data), nil
}
