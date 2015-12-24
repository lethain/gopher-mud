package mud

import (
	"bufio"
	"github.com/lethain/gopher-mud/player"
	"github.com/lethain/gopher-mud/telnet"
	"log"
	"net"
	"strings"
)

type MudServer struct {
	Loc          string
	TelnetServer *telnet.TelnetServer
}

func (ms *MudServer) ListenAndServe() {
	// setup connectino to database
	player.InitDatabase()

	// setup and run server
	telnet := telnet.TelnetServer{Loc: ms.Loc, Handler: ms.HandleConn}
	ms.TelnetServer = &telnet
	telnet.ListenAndServe()
}

func (ms *MudServer) HandleConn(conn net.Conn) {
	defer conn.Close()
	p := player.NewPlayer(conn)
	log.Printf("[%v]\tNew connection from %v", p.ShortID(), conn.RemoteAddr())

	go func() {
		for {
			msg, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				log.Printf("[%v]\tError reading line: %v", p.ShortID(), err)
				continue
			}
			msg = strings.Trim(msg, "\n\r\t ")
			resp, err := p.HandleMessage(msg)

			if resp != "" {
				p.Msgs <- resp
			}

			if err != nil {
				if err != player.Exited {
					log.Printf("[%v]\tError handling message: %v", p.ShortID(), err)
				}
				p.Msgs <- "We encountered an error and are logging you out."
				p.Logout()
				return
			}
		}
	}()

	conn.Write([]byte(p.Start()))
	for msg := range p.Msgs {
		// always end with newline
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		conn.Write([]byte(msg))
	}

}
