package player

import (
	"fmt"
	"log"
	"strings"
)

type CmdFunc func(*Player, string) (string, error)

type Command struct {
	Name    string
	Help    string
	Aliases []string
	Func    CmdFunc
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

func NewQuitCmd() *Command {
	return &Command{
		Name:    "quit",
		Aliases: []string{"exit", "q"},
		Func: func(p *Player, cmd string) (string, error) {
			return "See you next time.", Exited
		},
	}
}

func LoginCmd() *Command {
	return &Command{
		Name:    "login",
		Aliases: []string{"l"},
		Func: func(p *Player, cmd string) (string, error) {
			return p.SwitchModes(LoginUsernameMode, cmd), nil
		},
	}

}

func GetUsernameFunc(p *Player, cmd string) (string, error) {
	player, ok := GetPlayer(cmd)
	if ok == false {
		return fmt.Sprintf("Player with name %v doesn't exist yet. [Create] to go to character creation.\n%v", cmd, p.Mode.Render()), nil
	}
	p.MergePlayer(player)
	return p.SwitchModes(LoginPasswordMode), nil
}
