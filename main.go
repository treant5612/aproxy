package main

import (
	"aproxy/socks5"
)

func main() {
	sok := socks5.NewServer()
	sok.ListenAndServe("tcp",":1099")

}
