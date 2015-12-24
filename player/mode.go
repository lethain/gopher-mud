package player

import (
	"bytes"
	"sync"
	"errors"
	"fmt"
	"log"
	"text/template"
)

var UnknownMode = errors.New("No such mode exists.")
type InitModeFunc func(*Player)

const (
	SplashMode = iota
	LoginUsernameMode
	LoginPasswordMode
	GameMode
	CreateCharacterMode
)

type Mode struct {
	Id         int
	Name       string
	Desc       string
	DescTemplate   string
	Cmds       []*Command
	DefaultCmd CmdFunc
	InitCmd    InitModeFunc
}

var ModeTemplateCache = map[string]*template.Template{}

func (m *Mode) String() string {
	return fmt.Sprintf("Mode(%v)", m.Name)
}

func (m *Mode) Render(p *Player) string {
	if m.Desc == "" && m.DescTemplate != "" {
		tmpl, exists := ModeTemplateCache[m.DescTemplate]
		if !exists {
			newTmpl, err := template.ParseFiles(m.DescTemplate)
			if err != nil {
				log.Printf("error loading template: %v", err)
				return m.Desc
			}
			ModeTemplateCache[m.DescTemplate] = newTmpl
			tmpl = newTmpl
		}
		var rendered bytes.Buffer
		err := tmpl.Execute(&rendered, p)
		if err != nil {
			log.Printf("error rendering template: %v", err)
		}
		return rendered.String()
	}
	return m.Desc
}

var modes map[int]*Mode

func NewSplashMode() *Mode {
	mode := Mode{Id: SplashMode, Name: "Splash", DescTemplate: "splash.txt"}
	mode.Cmds = []*Command{LoginCmd(), CreateCharacterCmd(), NewQuitCmd()}
	return &mode
}

func NewCreateCharacterMode() *Mode {
	mode := Mode{Id: CreateCharacterMode, Name: "CreateCharacterMode", DescTemplate: "create_character.txt"}
	mode.Cmds = []*Command{NewQuitCmd()}
	mode.DefaultCmd = CreateCharacterFunc
	return &mode
}

func NewLoginUsernameMode() *Mode {
	mode := Mode{Id: SplashMode, Name: "LoginUsername", DescTemplate: "login_username.txt"}
	mode.Cmds = []*Command{NewQuitCmd()}
	mode.DefaultCmd = GetUsernameFunc
	return &mode
}

func NewLoginPasswordMode() *Mode {
	mode := Mode{Id: LoginPasswordMode, Name: "LoginPassword", DescTemplate: "login_password.txt"}
	mode.Cmds = []*Command{NewQuitCmd()}
	mode.DefaultCmd = GetPasswordFunc
	return &mode
}

var GameState = struct{
	sync.RWMutex
	Players map[string]*Player
}{Players: make(map[string]*Player)}


func NewGameMode() *Mode {
	mode := Mode{Id: GameMode, Name: "GamePassword", Desc: " GameMode!!!!!!"}
	mode.Cmds = []*Command{NewQuitCmd(), StatusCmd(), WhoCmd()}
	mode.InitCmd = func(p *Player) {
		GameState.Lock()
		GameState.Players[p.Name] = p
		GameState.Unlock()
	}
	return &mode
}

func GetMode(mode int) (*Mode, error) {
	switch mode {
	case SplashMode:
		return NewSplashMode(), nil
	case GameMode:
		return NewGameMode(), nil
	case LoginUsernameMode:
		return NewLoginUsernameMode(), nil
	case LoginPasswordMode:
		return NewLoginPasswordMode(), nil
	case CreateCharacterMode:
		return NewCreateCharacterMode(), nil
	default:
		return &Mode{}, UnknownMode
	}
}

func MustGetMode(mode int) *Mode {
	m, err := GetMode(mode)
	if err != nil {
		panic(err)
	}
	return m
}
