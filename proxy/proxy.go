package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

func Proxy(from, to net.Conn) error {

	done := make(chan error, 2)

	go func() {
		_, err := io.Copy(from, to)

		done <- err

	}()

	go func() {
		_, err := io.Copy(to, from)

		done <- err

	}()

	err1 := <-done

	if err1 != nil {
		return err1
	}

	err2 := <-done

	return err2
}

func ProxyServer(addr, serverAddr string) {
	Listener, err := net.Listen("tcp", addr)

	if err != nil {
		fmt.Printf("Proxy server listening issue, %s", err)
	}

	log.Printf("Proxy Server listening at address %s at time %s", addr, time.Now())

	for {
		conn, err := Listener.Accept()

		if err != nil {
			fmt.Printf("Proxy server Conn Accepting issue, %s", err)
		}

		d := net.Dialer{}

		go handleConnection(conn, serverAddr, d)
	}
}

func handleConnection(from net.Conn, serverAddr string, d net.Dialer) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))

	defer func() {
		cancel()
		from.Close()
	}()

	to, err := d.DialContext(ctx, "tcp", serverAddr)

	if err != nil {
		fmt.Printf("Error connecting to server %q", err)
		return
	}

	defer to.Close()

	err = Proxy(from, to)

	if err != nil {
		fmt.Print("Proxy err", err)
		return
	}

}
