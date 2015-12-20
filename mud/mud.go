package mud

import (
	"bufio"
	"strings"
	"log"
	"net"
	"github.com/lethain/gopher-mud/telnet"
	"github.com/lethain/gopher-mud/player"
)


type MudServer struct {
	Loc string
	TelnetServer *telnet.TelnetServer
}

func (ms *MudServer) ListenAndServe() {
	// load all modes
	player.LoadModes()

	// setup and run server
	telnet := telnet.TelnetServer{Loc: ms.Loc, Handler: ms.HandleConn}
	ms.TelnetServer = &telnet
	telnet.ListenAndServe()
}

func (ms *MudServer)  HandleConn(conn net.Conn) {
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