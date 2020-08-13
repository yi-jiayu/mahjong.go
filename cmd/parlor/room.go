package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"math/rand"
	"sync"

	"github.com/yi-jiayu/mahjong.go"
)

const (
	PhaseLobby = iota
	PhaseInProgress
	PhaseRoundOver
)

type Room struct {
	ID      string
	Nonce   int
	Phase   int
	Players []string
	Round   *mahjong.Round

	sync.RWMutex
	clients map[chan string]struct{}
}

type Participant struct {
	Seat      int            `json:"seat"`
	Concealed []mahjong.Tile `json:"concealed,omitempty"`
}

func (r *Room) UnmarshalBinary(data []byte) error {
	rdr := bytes.NewReader(data)
	dec := gob.NewDecoder(rdr)
	var room struct {
		Nonce   int
		Phase   int
		Players []string
		Round   *mahjong.Round
	}
	err := dec.Decode(&room)
	if err != nil {
		return err
	}
	r.Nonce = room.Nonce
	r.Phase = room.Phase
	r.Players = room.Players
	r.Round = room.Round
	r.clients = map[chan string]struct{}{}
	return nil
}

func (r *Room) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(struct {
		Nonce   int
		Phase   int
		Players []string
		Round   *mahjong.Round
	}{
		Nonce:   r.Nonce,
		Phase:   r.Phase,
		Players: r.Players,
		Round:   r.Round,
	})
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (r *Room) MarshalJSON() ([]byte, error) {
	players := make([]string, len(r.Players))
	for i, playerID := range r.Players {
		players[i] = playerRegistry.GetName(playerID)
	}
	return json.Marshal(struct {
		Nonce   int            `json:"nonce"`
		Phase   int            `json:"phase"`
		Players []string       `json:"players"`
		Round   *mahjong.Round `json:"round"`
	}{
		Nonce:   r.Nonce,
		Phase:   r.Phase,
		Players: players,
		Round:   r.Round,
	})
}

func NewRoom(host string) *Room {
	return &Room{
		Players: []string{host},
		clients: map[chan string]struct{}{},
	}
}

func (r *Room) AddClient(c chan string) {
	r.Lock()
	defer r.Unlock()

	r.clients[c] = struct{}{}
	var b bytes.Buffer
	json.NewEncoder(&b).Encode(r)
	c <- b.String()
}

func (r *Room) AddPlayer(id string) error {
	r.Lock()
	defer r.Unlock()
	if len(r.Players) == 4 {
		return errors.New("room full")
	}
	for _, p := range r.Players {
		if p == id {
			return errors.New("already joined")
		}
	}
	r.Players = append(r.Players, id)
	r.broadcast()
	return nil
}

func (r *Room) RemoveClient(c chan string) {
	r.Lock()
	defer r.Unlock()

	delete(r.clients, c)
}

func (r *Room) GetParticipant(playerID string) Participant {
	r.RLock()
	defer r.RUnlock()
	for i, id := range r.Players {
		if id == playerID {
			var concealed []mahjong.Tile
			if r.Round != nil {
				concealed = r.Round.Hands[i].Concealed
			}
			return Participant{
				Seat:      i,
				Concealed: concealed,
			}
		}
	}
	return Participant{
		Seat: -1,
	}
}

func (r *Room) broadcast() {
	var b bytes.Buffer
	json.NewEncoder(&b).Encode(r)
	for c := range r.clients {
		c <- b.String()
	}
}

func (r *Room) startRound() error {
	if len(r.Players) < 4 {
		return errors.New("not enough players")
	}
	r.Round = mahjong.NewRound(rand.Int63())
	r.Phase = PhaseInProgress
	return nil
}

type Action struct {
	Nonce int              `json:"nonce"`
	Type  string           `json:"type"`
	Tiles []mahjong.Tile   `json:"tiles"`
	Melds [][]mahjong.Tile `json:"melds"`
}

type DrawResult struct {
	Drawn   mahjong.Tile   `json:"drawn"`
	Flowers []mahjong.Tile `json:"flowers"`
}

func (r *Room) handleAction(playerID string, action Action) (interface{}, error) {
	seat := -1
	for i, p := range r.Players {
		if p == playerID {
			seat = i
			break
		}
	}
	if seat == -1 {
		return nil, errors.New("player not in room")
	}
	if action.Nonce != r.Nonce {
		return nil, errors.New("invalid nonce")
	}
	switch action.Type {
	case "start":
		return nil, r.startRound()
	case "discard":
		if len(action.Tiles) < 0 {
			return nil, errors.New("not enough tiles")
		}
		return nil, r.Round.Discard(mahjong.Direction(seat), action.Tiles[0])
	case "draw":
		drawn, flowers, err := r.Round.Draw(mahjong.Direction(seat))
		if err != nil {
			return nil, err
		}
		return DrawResult{
			Drawn:   drawn,
			Flowers: flowers,
		}, nil
	case "chow":
		if len(action.Tiles) < 2 {
			return nil, errors.New("not enough tiles")
		}
		return nil, r.Round.Chow(mahjong.Direction(seat), action.Tiles[0], action.Tiles[1])
	case "peng":
		if len(action.Tiles) < 1 {
			return nil, errors.New("not enough tiles")
		}
		return nil, r.Round.Peng(mahjong.Direction(seat), action.Tiles[0])
	case "kong":
		if len(action.Tiles) < 1 {
			return nil, errors.New("not enough tiles")
		}
		drawn, flowers, err := r.Round.Kong(mahjong.Direction(seat), action.Tiles[0])
		if err != nil {
			return nil, err
		}
		return DrawResult{
			Drawn:   drawn,
			Flowers: flowers,
		}, nil
	case "hu":
		err := r.Round.Win(mahjong.Direction(seat), action.Melds)
		if err != nil {
			return nil, err
		}
		r.Phase = PhaseRoundOver
		return nil, nil
	default:
		return nil, errors.New("invalid action")
	}
}

func (r *Room) HandleAction(playerID string, action Action) (interface{}, error) {
	r.Lock()
	defer r.Unlock()
	result, err := r.handleAction(playerID, action)
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = struct{}{}
	}
	r.Nonce++
	r.broadcast()
	return result, nil
}
