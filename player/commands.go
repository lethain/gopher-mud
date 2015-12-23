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

func CreateCharacterCmd() *Command {
	return &Command{
		Name:    "create",
		Aliases: []string{"c"},
		Func: func(p *Player, cmd string) (string, error) {
			return p.SwitchModes(CreateCharacterMode, cmd), nil
		},
	}
}

func GetUsernameFunc(p *Player, cmd string) (string, error) {
	player, ok := GetPlayer(cmd)
	if ok == false {
		return fmt.Sprintf("Player with name %v doesn't exist yet. [Create] to go to character creation.\n%v", cmd, p.Mode.Render(p)), nil
	}
	p.MergePlayer(player)
	return p.SwitchModes(LoginPasswordMode), nil
}

func GetPasswordFunc(p *Player, cmd string) (string, error) {
	if !p.CheckPassword(cmd) {
		return fmt.Sprintf("Sorry %v, couldn't recognize your password.", p.Name), nil
	}
	return p.SwitchModes(GameMode), nil
}


func CreateCharacterFunc(p *Player, cmd string) (string, error) {
	log.Printf("CreateCharacter: %v, %v, %v", p.Mode, p, cmd)
	if p.Name == "" {
		if len(cmd) < 4 {
			return "Name must be at least four letters long.", nil
		}
		_, ok := GetPlayer(cmd)
		if ok == true {
			return fmt.Sprintf("The name %v is already taken.", cmd), nil
		}
		p.Name = cmd
	}

	if p.Name != "" {
		if err := p.Save(); err != nil {
			log.Printf("error saving %v: %v", p, err)
			return "Couldn't create your new character.", err
		}
		return p.SwitchModes(GameMode), nil
	}
	return p.Mode.Render(p), nil
}
