package player

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"net"
	"strings"
)

var Exited = errors.New("Player exited.")

type Race int
const (
	RaceNone  = iota
	RaceEarther
	RaceLunite
	RaceBelter
)

type Player struct {
	Conn     net.Conn
	LoggedIn bool
	Name     string
	UUID     string
	Level    int
	Exp      int
	Race     Race
	Mode     *Mode
	HP       int
	MaxHP    int
	SP       int
	Msgs     chan string
}

func (p *Player) RaceString() string {
	switch p.Race {
	case RaceEarther:
		return "Earther"
	case RaceLunite:
		return "Lunite"
	case RaceBelter:
		return "Belter"
	default:
		return "Unknown"
	}
}

func NewPlayer(conn net.Conn) *Player {
	return &Player{Conn: conn, UUID: uuid.NewV4().String(), LoggedIn: false, Msgs: make(chan string, 10)}
}

var db *leveldb.DB

func InitDatabase() {
	var err error
	db, err = leveldb.OpenFile("path/to/db", nil)
	if err != nil {
		panic(err)
	}
}

func GetPlayer(name string) (*Player, bool) {
	data, err := db.Get([]byte(name), nil)
	if err != nil {
		if err != leveldb.ErrNotFound {
			log.Printf("unexpected error retrieving player: %v", err)
		}
		return &Player{}, false
	}
	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)
	var newPlayer Player
	err = dec.Decode(&newPlayer)
	if err != nil {
		return &newPlayer, false
	}
	return &newPlayer, true
}

func (p *Player) Logout() {
	p.Save()
	GameState.Lock()
	delete(GameState.Players, p.Name)
	close(p.Msgs)
	p.Msgs = nil
	GameState.Unlock()

	log.Printf("[%v]\tLogged out.", p.ShortID())
}

func (p *Player) Save() error {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)

	// can't GOB the connection object, so doing this silly
	// unset and reset maneuver
	conn := p.Conn
	p.Conn = nil
	msgs := p.Msgs
	p.Msgs = nil
	defer func() {
		p.Conn = conn
		p.Msgs = msgs
	}()
	if err := enc.Encode(p); err != nil {
		return err
	}
	return db.Put([]byte(p.Name), buffer.Bytes(), nil)
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
	p.Name = op.Name
	p.UUID = op.UUID
	p.Level = op.Level
	p.Exp = op.Exp
	p.Race = op.Race
	p.HP = op.HP
	p.MaxHP = op.MaxHP
	p.SP = op.SP
}

func (p *Player) SwitchModes(mode int, cmd ...string) string {
	p.Mode = MustGetMode(mode)
	if p.Mode.InitCmd != nil {
		p.Mode.InitCmd(p)
	}
	log.Printf("[%v]\tSwitching to mode %v due to %v", p.ShortID(), p.Mode.Name, cmd)
	return p.Mode.Render(p)
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
	msg = strings.Trim(msg, " \t\n\r")
	words := strings.Split(strings.ToLower(msg), " ")
	if len(msg) > 0 && len(words) > 0 {
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
