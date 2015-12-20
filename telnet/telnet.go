package telnet

import (
	"log"
	"net"
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

