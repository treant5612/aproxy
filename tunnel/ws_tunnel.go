package tunnel

import (
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
)

type WebsocketServer struct {
	*commonServer
	pattern string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewWebsocketServer(key string, pattern string) (*WebsocketServer, error) {
	return &WebsocketServer{&commonServer{Key: key}, pattern}, nil
}

func (s *WebsocketServer) ListenWs(address string) {
	if s.pattern == "" {
		s.pattern = "/"
	}
	if address == "" {
		address = "0.0.0.0:80"
	}
	http.HandleFunc(s.pattern, s.handleFunc)
	http.ListenAndServe(address, nil)
}

func (s *WebsocketServer) handleFunc(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	rw := &wsReadWriterCloser{conn, nil}
	err = s.handle(rw)
	if err != nil {
		log.Println(err)
	}
}

type wsReadWriterCloser struct {
	conn    *websocket.Conn
	readBuf []byte
}

func (rw *wsReadWriterCloser) Close() error {
	return rw.conn.Close()
}

func (rw *wsReadWriterCloser) Read(buf []byte) (n int, err error) {
	if len(rw.readBuf) == 0 {
		_, msg, err := rw.conn.ReadMessage()
		if err != nil {
			return 0, err
		}
		rw.readBuf = msg
	}
	n = copy(buf, rw.readBuf)
	rw.readBuf = rw.readBuf[n:]
	return n, nil
}

func (rw *wsReadWriterCloser) Write(buf []byte) (n int, err error) {
	return len(buf), rw.conn.WriteMessage(websocket.BinaryMessage, buf)
}

type WebSocketClient struct {
	io.ReadWriteCloser
	*commonClient
}

func NewWebSocketClient(key, address, dstAddr, dstPort string) (wsClient *WebSocketClient, err error) {
	conn, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rw := &wsReadWriterCloser{conn: conn}
	return &WebSocketClient{rw, &commonClient{Key: key, dstAddr: dstAddr, dstPort: dstPort}}, nil
}
func (wsc *WebSocketClient) Transport(writer io.ReadWriter) error {
	return wsc.doTransport(wsc.ReadWriteCloser, writer)
}
