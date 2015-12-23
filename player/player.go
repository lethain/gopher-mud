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

type Player struct {
	Conn     net.Conn
	LoggedIn bool
	Name     string
	UUID     string
	Mode     *Mode
	HP       int
}

func NewPlayer(conn net.Conn) *Player {
	return &Player{Conn: conn, UUID: uuid.NewV4().String(), LoggedIn: false}
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
	if err == leveldb.ErrNotFound {
		/* Write some data for testing purposes... */
		// this should be replaced by character creation
		p := &Player{Name: name, UUID: uuid.NewV4().String()}
		if err := p.Save(); err != nil {
			log.Printf("error saving %v: %v", p, err)
		}
		return &Player{}, false
	} else if err != nil {
		log.Printf("unexpected error retrieving player: %v", err)
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

func (p *Player) Save() error {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
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
	p.UUID = op.UUID
	p.Name = op.Name
	p.HP = op.HP
}

func (p *Player) SwitchModes(mode int, cmd ...string) string {
	p.Mode = MustGetMode(mode)
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
