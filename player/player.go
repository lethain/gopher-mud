package player

import (
	"errors"
	"strings"
	"net"
	"log"
	"io/ioutil"
)

var  Exited = errors.New("Player exited.")
var UnknownMode = errors.New("No such mode exists.")

const (
	SplashMode = iota
	LoginUsernameMode
	LoginPasswordMode
	NormalMode
)

type Command struct {
	Name string
	Help string
	Aliases []string
	Func func(*Player, string) (string, error)
}

type Mode struct {
	Id int
	Name string
	Desc string
	DescFile string
	Cmds []*Command
	DefaultCmd *Command
}

func (m *Mode) Render() string {
	if m.Desc == "" && m.DescFile != "" {
		raw, err := ioutil.ReadFile(m.DescFile)
		if err != nil {
			log.Printf("failed to read splash.txt: %v", err)
			return m.Desc
		}
		m.Desc = string(raw)
	}
	return m.Desc
}

var modes map[int]*Mode

func LoadModes() {
	modes = map[int]*Mode{}
	modes[SplashMode] = &Mode{Id: SplashMode, Name: "Splash", DescFile: "splash.txt"}
	modes[LoginUsernameMode] = &Mode{Id: LoginUsernameMode, Name: "LoginUsername", DescFile: "login_username.txt"}
	modes[LoginPasswordMode] = &Mode{Id: LoginPasswordMode, Name: "LoginPassword", DescFile: "login_password.txt"}


	splashCmds := make([]*Command, 0)
	loginCmd := &Command{
		Name: "login",
		Aliases: []string{"l"},
		Func: func(p *Player, cmd string) (string, error) {
			return p.SwitchModes(LoginUsernameMode, cmd), nil
		},
	}
	splashCmds = append(splashCmds, loginCmd)

	modes[SplashMode].Cmds = splashCmds
}

func GetMode(mode int) (*Mode, error) {
	if modes[mode] == nil {
		return &Mode{}, UnknownMode
	}
	return modes[mode], nil
}

func MustGetMode(mode int) *Mode {
	m, err := GetMode(mode)
	if err != nil {
		panic(err)
	}
	return m
}


type Player struct {
	Conn net.Conn
	Created bool
	Name string
	UUID string
	Mode *Mode
	HP int
}

func (p *Player) SwitchModes(mode int, cmd ...string) string {
	p.Mode = MustGetMode(mode)
	log.Printf("Switching to mode %v due to %v", p.Mode.Name, cmd)
	return p.Mode.Render()
}


var id = 0

func (p *Player) HandleMessage(msg string) (string, error) {
	id++
	log.Printf("[%v]\tMsg Received (len %v): '%v'", id, len(msg), msg)
	if msg == "exit" {
		return "See you next time.", Exited
	}
	words := strings.Split(strings.ToLower(msg), " ")
	if len(words) > 0 {
		first := words[0]
		log.Printf("Trying to match %v", first)
		for _, cmd := range p.Mode.Cmds {
			log.Printf("[%v]\t(Full) Trying to match %v to %v", id, first, cmd.Name)
			if first == cmd.Name {
				return cmd.Func(p, msg)
			}
			for _, alias := range cmd.Aliases {
				log.Printf("[%v]\t(Alias) Trying to match %v to %v", id, first, alias)
				if first == alias {
					return cmd.Func(p, msg)
				}
			}
		}
	}
	return "", nil
}

func (p *Player) Splash() string {
	return p.SwitchModes(SplashMode)
}
