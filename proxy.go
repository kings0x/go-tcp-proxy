package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

func Proxy(from io.Reader, to io.Writer) error {
	fromIsWriter, ok := from.(io.Writer)
	if !ok {
		fmt.Print("from is not writer")

	}
	toIsReader, ok := to.(io.Reader)

	if !ok {
		fmt.Print("to is not writer")

	}

	go func() {
		_, _ = io.Copy(fromIsWriter, toIsReader)

	}()

	_, err := io.Copy(to, from)

	return err
}

func ProxyServer(addr, serverAddr string) {
	Listener, err := net.Listen("tcp", addr)

	if err != nil {
		fmt.Printf("Proxy server listening issue, %s", err)
	}

	for {
		conn, err := Listener.Accept()

		if err != nil {
			fmt.Printf("Proxy server Conn Accepting issue, %s", err)
		}

		go func(from net.Conn) {

			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))

			defer func() {
				cancel()
				conn.Close()

			}()

			d := net.Dialer{}

			to, err := d.DialContext(ctx, "tcp", serverAddr)

			if err != nil {
				fmt.Printf("Error connecting to server %q", err)
				return
			}

			err = Proxy(from, to)

			if err != nil {
				fmt.Print("Proxy err", err)
				return
			}
		}(conn)
	}
}
