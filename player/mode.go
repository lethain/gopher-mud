package player

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
)

var UnknownMode = errors.New("No such mode exists.")

const (
	SplashMode = iota
	LoginUsernameMode
	LoginPasswordMode
	GameMode
)

type Mode struct {
	Id         int
	Name       string
	Desc       string
	DescFile   string
	Cmds       []*Command
	DefaultCmd CmdFunc
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

func NewSplashMode() *Mode {
	mode := Mode{Id: SplashMode, Name: "Splash", DescFile: "splash.txt"}
	loginCmd := &Command{
		Name:    "login",
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
