package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type TelnetServer struct {
	Loc      string
	Listener net.Listener
	Handler  func(net.Conn)
}

func (t *TelnetServer) ListenAndServe() {
	log.Printf("Starting TelnetServer on %v", t.Loc)
	ln, err := net.Listen("tcp", t.Loc)
	if err != nil {
		panic(err)
	}
	t.Listener = ln
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v", err)
			continue
		}
		go t.Handler(conn)
	}

	select {}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	log.Printf("handleConn: %v", conn)
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Printf("error reading line: %v", err)
			continue
		}
		if message == "exit" {
			return
		}
		log.Printf("Message Received (len %v): '%v'", len(message), message)

		newmessage := strings.ToUpper(message)
		conn.Write([]byte(fmt.Sprintf("(%v)\t%v\n", len(message), newmessage)))
	}

}

func main() {
	telnet := TelnetServer{Loc: ":9000", Handler: handleConn}
	telnet.ListenAndServe()
}
