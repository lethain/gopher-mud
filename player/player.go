package player

import (
	"errors"
	"fmt"
	"net"
	"log"
	"io/ioutil"
)

var  Exited = errors.New("Player exited.")


type Player struct {
	Conn net.Conn
	Created bool
	Name string
	UUID string
	HP int
}

func (p *Player) HandleMessage(msg string) (string, error) {
	log.Printf("Msg Received (len %v): '%v'", len(msg), msg)

	if msg == "exit" {
		return "See you next time.", Exited
	}

	return fmt.Sprintf("(%v) Received: \t%v", len(msg), msg), nil
}


var splashCache string

func (p *Player) Splash() string {
	if splashCache == "" {
		raw, err := ioutil.ReadFile("splash.txt")
		if err != nil {
			log.Printf("failed to read splash.txt: %v", err)
		}
		splashCache = string(raw)
	}
	return splashCache
}
