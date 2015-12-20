package main

import (
	"github.com/lethain/gopher-mud/telnet"
	"github.com/lethain/gopher-mud/player"	
	"bufio"
	"strings"
	"log"
	"net"
	"flag"
)

var loc = flag.String("loc", ":9000", "location:port to run server")


func handleConn(conn net.Conn) {
	defer conn.Close()
	p := player.Player{Conn: conn}
	log.Printf("new connection from %v", p)

	conn.Write([]byte(p.Splash()))
	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Printf("error reading line: %v", err)
			continue
		}
		msg = strings.Trim(msg, "\n\r\t ")
		resp, err := p.HandleMessage(msg)

		// always end with newline
		if !strings.HasSuffix(resp, "\n") {
			resp += "\n"
		}
		
		conn.Write([]byte(resp))
		if err != nil {
			log.Printf("error handling message: %v", err)
			return
		}
	}
}

func main() {
	flag.Parse()
	telnet := telnet.TelnetServer{Loc: *loc, Handler: handleConn}
	telnet.ListenAndServe()
}
