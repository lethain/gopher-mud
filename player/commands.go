package player

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"
)

type CmdFunc func(*Player, string) (string, error)
var CommandTemplateCache = map[string]*template.Template{}


func RenderCommandTemplate(path string, p interface{}) (string, error) {
	tmpl, exists := CommandTemplateCache[path]
	if !exists {
		newTmpl, err := template.ParseFiles(path)
		if err != nil {
			log.Printf("error loading template: %v", err)
			return "", err
		}
		CommandTemplateCache[path] = newTmpl
		tmpl = newTmpl
	}
	var rendered bytes.Buffer
	err := tmpl.Execute(&rendered, p)
	if err != nil {
		log.Printf("error rendering template: %v", err)
		return "", err
	}
	return rendered.String(), nil
}

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
			GameState.Lock()
			delete(GameState.Players, p.Name)
			GameState.Unlock()	
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

func StatusCmd() *Command {
	return &Command{
		Name: "status",
		Func: func(p *Player, cmd string) (string, error) {
			return RenderCommandTemplate("cmd_status.txt", p)
		},
	}
}

func WhoCmd() *Command {
	return &Command{
		Name: "who",
		Func: func(p *Player, cmd string) (string, error) {
			GameState.RLock()
			defer GameState.RUnlock()
			return RenderCommandTemplate("cmd_who.txt", GameState)
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
	} else if p.Race == RaceNone {
		switch strings.ToLower(cmd) {
		case "earther":
			p.Race = RaceEarther
		case "lunite":
			p.Race = RaceLunite
		case "belter":
			p.Race = RaceBelter
		default:
			return fmt.Sprintf("Race %v didn't match a valid option:  earther, lunite or belter.", cmd), nil
		}
	}

	if p.Name != "" && p.Race != RaceNone {
		p.Level = 1
		p.HP = 100
		p.MaxHP = 100
		p.SP = 0
		p.Exp = 0
		if err := p.Save(); err != nil {
			log.Printf("error saving %v: %v", p, err)
			return "Couldn't create your new character.", err
		}
		return p.SwitchModes(GameMode), nil
	}
	return p.Mode.Render(p), nil
}
