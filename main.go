package main

import "github.com/kings0x/go-tcp-proxy/proxy"

func main() {
	addr := "127.0.0.1:1"
	serverAddr := "127.0.0.1:80"

	go StartListener(serverAddr)
	go proxy.ProxyServer(addr, serverAddr)
	Startconn(addr)
}
