package player

import (
	"fmt"
	"errors"
	"strings"
	"net"
	"log"
	"io/ioutil"
	"github.com/satori/go.uuid"
)

var  Exited = errors.New("Player exited.")
var UnknownMode = errors.New("No such mode exists.")

type CmdFunc func(*Player, string) (string, error)

const (
	SplashMode = iota
	LoginUsernameMode
	LoginPasswordMode
	GameMode
)

type Command struct {
	Name string
	Help string
	Aliases []string
	Func CmdFunc
}

type Mode struct {
	Id int
	Name string
	Desc string
	DescFile string
	Cmds []*Command
	DefaultCmd CmdFunc
}

func NoValidCommand(p *Player, msg string) (string, error) {
	allowed := make([]string, 0)
	for _, cmd := range p.Mode.Cmds {
		allowed = append(allowed, cmd.Name)
	}
	allowedCmds := strings.Join(allowed, ", ")
	log.Printf("[%v]\tCommand %v didn't match any of %v.", p.ShortID(), msg, allowedCmds)
	return fmt.Sprintf("Sorry, didn't recognize that command. Try one of %v.", allowedCmds), nil
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



func NewQuitCmd() *Command {
	return &Command{
		Name :"quit",
		Aliases: []string{"exit", "q"},
		Func: func(p *Player, cmd string) (string, error) {
			return "See you next time.", Exited
		},
	}
}

func NewSplashMode() *Mode {
	mode := Mode{Id: SplashMode, Name: "Splash", DescFile: "splash.txt"}
	loginCmd := &Command{
		Name: "login",
		Aliases: []string{"l"},
		Func: func(p *Player, cmd string) (string, error) {
			return p.SwitchModes(LoginUsernameMode, cmd), nil
		},
	}
	splashCmds := make([]*Command, 0)
	splashCmds = append(splashCmds, loginCmd, NewQuitCmd())
	mode.Cmds = splashCmds
	return &mode
}

func NewLoginUsernameMode() *Mode {
	mode := Mode{Id: SplashMode, Name: "LoginUsername", DescFile: "login_username.txt"}
	mode.Cmds = []*Command{NewQuitCmd()}
	mode.DefaultCmd = func(p *Player, cmd string) (string, error) {
		player, ok := GetPlayer(cmd)
		if ok == false {
			return fmt.Sprintf("Player with name %v doesn't exist yet. [Create] to go to character creation.\n%v", cmd, p.Mode.Render()), nil
		} else {
			p.MergePlayer(player)
			return p.SwitchModes(LoginPasswordMode), nil
		}
	}
	return &mode
}

func NewLoginPasswordMode() *Mode {
	mode := Mode{Id: LoginPasswordMode, Name: "LoginPassword", DescFile: "login_password.txt"}
	mode.Cmds = []*Command{NewQuitCmd()}
	mode.DefaultCmd = func(p *Player, cmd string) (string, error) {
		if !p.CheckPassword(cmd) {
			return fmt.Sprintf("Sorry %v, couldn't recognize your password.", p.Name), nil
		}
		return p.SwitchModes(GameMode), nil
	}
	return &mode
}

func NewGameMode() *Mode {
	mode := Mode{Id: GameMode, Name: "GamePassword", Desc: " GameMode!!!!!!"}
	mode.Cmds = []*Command{NewQuitCmd()}
	return &mode
}



func LoadModes() {
	modes = map[int]*Mode{}
	modes[SplashMode] = NewSplashMode()
	modes[GameMode] = NewGameMode()
	modes[LoginUsernameMode] = NewLoginUsernameMode()
	modes[LoginPasswordMode] = NewLoginPasswordMode()
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
	LoggedIn bool
	Name string
	UUID string
	Mode *Mode
	HP int
}

func NewPlayer(conn net.Conn) *Player {
	return &Player{Conn: conn, UUID: uuid.NewV4().String(), LoggedIn: false}
}

var PlayersByName map[string]*Player

func LoadPlayers() {
	PlayersByName = map[string]*Player{}
	PlayersByName["lethain"] = &Player{
		Name: "lethain",
		UUID: "9999",
		HP: 1000,
	}
}

func GetPlayer(name string) (*Player, bool) {
	player, ok := PlayersByName[name]
	return player, ok
}

func (p *Player) String() string {
	return fmt.Sprintf("Player(%v)", p.UUID[:4])
}

func (p *Player) ShortID() string {
	if p.Name != "" {
		return p.Name
	} else {
		return p.UUID[:4]
	}
}

func (p *Player) MergePlayer(op *Player) {
	log.Printf("[%v]\tID %v transitioning to ID %v.", p.ShortID(), p.ShortID(), op.ShortID())
	p.UUID = op.UUID
	p.Name = op.Name
	p.HP = op.HP
}

func (p *Player) SwitchModes(mode int, cmd ...string) string {
	p.Mode = MustGetMode(mode)
	log.Printf("[%v]\tSwitching to mode %v due to %v", p.ShortID(), p.Mode.Name, cmd)
	return p.Mode.Render()
}

func (p *Player) CheckPassword(pwd string) bool {
	// ok, so this clearly not a complete password check
	if len(pwd) > 5 {
		return true
	}
	return false
}


func (p *Player) HandleMessage(msg string) (string, error) {
	log.Printf("[%v]\tMsg Received (len %v): '%v'", p.ShortID(), len(msg), msg)
	words := strings.Split(strings.ToLower(msg), " ")
	if len(words) > 0 {
		first := words[0]
		for _, cmd := range p.Mode.Cmds {
			if first == cmd.Name {
				return cmd.Func(p, msg)
			}
			for _, alias := range cmd.Aliases {
				if first == alias {
					return cmd.Func(p, msg)
				}
			}
		}
		if p.Mode.DefaultCmd != nil {
			return p.Mode.DefaultCmd(p, msg)
		}
		return NoValidCommand(p, msg)
	}
	return "", nil
}

func (p *Player) Splash() string {
	return p.SwitchModes(SplashMode)
}
