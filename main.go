package main

import (
	"github.com/lethain/gopher-mud/telnet"
	"bufio"
	"strings"
	"fmt"
	"log"
	"net"
)


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
	telnet := telnet.TelnetServer{Loc: ":9000", Handler: handleConn}
	telnet.ListenAndServe()
}
