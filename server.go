package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func StartListener(addr string) {

	listner, err := net.Listen("tcp", addr)

	if err != nil {
		log.Println("listener did not listen", err)
	}

	log.Printf("Server listening at address %s at time %s", addr, time.Now())

	for {
		conn, err := listner.Accept()

		if err != nil {
			log.Print("Server: err in Accepting connection")
			return

		}

		go handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {

	defer c.Close()

	for {
		if err := c.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
				fmt.Print("time out error: ", err)
				return

			}

			fmt.Print("client: setReadDeadline err: ", err)
			return
		}

		payload, err := decode(c)

		if err != nil {
			log.Print("Server err with decoding", err)
			return
		}

		log.Printf("recieved message from client: %s", payload.String())

		message := Binary("Signed client message by server" + payload.String())

		if err = c.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
			if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
				log.Print("time out error: ", err)
				return
			}

			log.Print("client: setwriteDeadline err: ", err)
			return

		}

		_, err = message.WriteTo(c)

		if err != nil {
			log.Print("err writing to client", err)
		}

		time.Sleep(10 * time.Second)

		// _ = c.Close()

	}

}
