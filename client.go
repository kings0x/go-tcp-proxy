package main

import (
	"context"
	"fmt"
	"net"
	"time"
)

func Startconn(proxyAddr string) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer cancel()

	d := net.Dialer{}

	conn, err := d.DialContext(ctx, "tcp", proxyAddr)

	if err != nil {
		fmt.Printf("conn: failure to connect to listener %v\n", err)
		return
	}

	defer conn.Close()

	for {
		message := Binary("something fishy going on")

		if err = conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
			if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
				fmt.Print("time out error: ", err)
				return
			}

			fmt.Print("client: setwriteDeadline err: ", err)
			return

		}

		_, err = message.WriteTo(conn)

		if err != nil {
			fmt.Printf("conn: write err: %q", err)
			return
		}

		if err = conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
				fmt.Print("time out error: ", err)
				return

			}

			fmt.Print("client: setReadDeadline err: ", err)
			return
		}

		payload, err := decode(conn)

		if err != nil {
			fmt.Printf("conn: read err: %q", err)
			return
		}

		fmt.Println("Recieved this from server passed through a proxy", payload.String())

		time.Sleep(2 * time.Second)
	}

}
