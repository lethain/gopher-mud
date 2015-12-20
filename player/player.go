package player

import (
	"fmt"
	"errors"
	"strings"
	"net"
	"log"
	"github.com/satori/go.uuid"
)

var  Exited = errors.New("Player exited.")

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

func (p *Player) Start() string {
	return p.SwitchModes(SplashMode)
}
