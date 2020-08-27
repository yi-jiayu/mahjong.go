package parlour

import (
	"errors"
	"sync"

	"github.com/yi-jiayu/mahjong.go/mahjong"
)

type Phase int

const (
	PhaseLobby Phase = iota
	PhaseInProgress
)

type Player struct {
	id   string
	Name string
}

type Room struct {
	ID      string
	Nonce   int
	Phase   Phase
	Players []Player
	Game    *mahjong.Game

	sync.RWMutex
}

type RoomView struct {
	ID      string            `json:"id"`
	Nonce   int               `json:"nonce"`
	Phase   Phase             `json:"state"`
	Players []Player          `json:"players"`
	Game    *mahjong.GameView `json:"game"`
}

func (r *Room) WithLock(f func(r *Room)) {
	r.Lock()
	f(r)
	r.Unlock()
}

func (r *Room) WithRLock(f func(r *Room)) {
	r.RLock()
	f(r)
	r.RUnlock()
}

func (r *Room) seat(playerID string) int {
	for i, player := range r.Players {
		if player.id == playerID {
			return i
		}
	}
	return -1
}

// view returns a player's view of a room.
func (r *Room) view(playerID string) RoomView {
	view := RoomView{
		ID:      r.ID,
		Nonce:   r.Nonce,
		Phase:   r.Phase,
		Players: r.Players,
	}
	return view
}

func (r *Room) addPlayer(player Player) error {
	for _, p := range r.Players {
		if p.Name == player.Name {
			return errors.New("name already taken")
		}
	}
	if len(r.Players) == 4 {
		return errors.New("room full")
	}
	r.Players = append(r.Players, player)
	return nil
}

func NewRoom(host Player) *Room {
	room := &Room{
		Phase:   PhaseLobby,
		Players: []Player{host},
	}
	return room
}
