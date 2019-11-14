package main

import (
	"aproxy/socks5"
	"aproxy/transport"
	"log"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	go asRemoteTransporter()
	asLocalSocksServer()
	time.Sleep(1 * time.Hour)
}
func asRemoteTransporter() {
	server := &transport.RemoteTransporterServer{Key: "testkey"}
	err := server.ListenAndServe("tcp", ":4721")
	log.Println(err)
}
func asLocalSocksServer() {
	server := socks5.NewServer()
	server.ListenAndServe("tcp", ":10909")
}
