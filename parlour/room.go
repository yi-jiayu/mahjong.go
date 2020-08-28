package parlour

import (
	"bytes"
	"encoding/json"
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
	Name string `json:"name"`
}

type Room struct {
	ID      string
	Nonce   int
	Phase   Phase
	Players []Player
	Game    *mahjong.Game

	sync.RWMutex

	// clients is a map of subscription channels to player IDs.
	clients map[chan string]string
}

type RoomView struct {
	ID      string            `json:"id"`
	Nonce   int               `json:"nonce"`
	Phase   Phase             `json:"phase"`
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
			if p.id == player.id {
				return nil
			}
			return errors.New("name already taken")
		}
	}
	if len(r.Players) == 4 {
		return errors.New("room full")
	}
	r.Players = append(r.Players, player)
	r.broadcast()
	return nil
}

func (r *Room) start(playerID string) error {
	return nil
}

type Action struct {
	Type  string         `json:"type"`
	Tiles []mahjong.Tile `json:"tiles"`
}

func (r *Room) update(playerID string, action Action) error {
	seat := r.seat(playerID)
	if seat == -1 {
		return errors.New("forbidden")
	}
	//t := time.Now()
	switch action.Type {

	}
	return nil
}

// AddClient subscribes a new client to the room. The current room state will
// be immediately sent through ch, so either ensure ch is buffered or read from
// ch concurrently to prevent deadlock.
func (r *Room) AddClient(playerID string, ch chan string) {
	r.Lock()
	defer r.Unlock()
	r.clients[ch] = playerID
	var b bytes.Buffer
	json.NewEncoder(&b).Encode(r.view(playerID))
	ch <- b.String()
}

func (r *Room) RemoveClient(ch chan string) {
	r.Lock()
	delete(r.clients, ch)
	r.Unlock()
}

func (r *Room) broadcast() {
	for ch, playerID := range r.clients {
		var b bytes.Buffer
		json.NewEncoder(&b).Encode(r.view(playerID))
		ch <- b.String()
	}
}

func (r *Room) removePlayer(playerID string) {
	for i, player := range r.Players {
		if player.id == playerID {
			r.Players = append(r.Players[:i], r.Players[i+1:]...)
			r.broadcast()
			return
		}
	}
}

func NewRoom(host Player) *Room {
	room := &Room{
		Phase:   PhaseLobby,
		Players: []Player{host},
		clients: make(map[chan string]string),
	}
	return room
}
