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
)

type Room struct {
	Nonce   int
	Phase   int
	Players []string
	Round   *mahjong.Round

	sync.Mutex
	clients map[chan string]struct{}
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

func (r *Room) addClient(c chan string) {
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
	return nil
}

func (r *Room) removeClient(c chan string) {
	r.Lock()
	defer r.Unlock()

	delete(r.clients, c)
}

func (r *Room) broadcast() {
	r.Lock()
	defer r.Unlock()

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
	r.Round = mahjong.NewRound(rand.Int63(), mahjong.DirectionEast)
	r.Phase = PhaseInProgress
	r.Nonce++
	return nil
}

type Action struct {
	Nonce int      `json:"nonce"`
	Type  string   `json:"type"`
	Tiles []string `json:"tiles"`
}

func (r *Room) HandleAction(playerID string, action Action) error {
	r.Lock()
	defer r.Unlock()
	seat := -1
	for i, p := range r.Players {
		if p == playerID {
			seat = i
			break
		}
	}
	if seat == -1 {
		return errors.New("player not in room")
	}
	if action.Nonce != r.Nonce {
		return errors.New("invalid nonce")
	}
	switch action.Type {
	case "start":
		return r.startRound()
	case "discard":
		if len(action.Tiles) < 0 {
			return errors.New("not enough tiles")
		}
		return r.Round.Discard(mahjong.Direction(seat), action.Tiles[0])
	case "draw":
		return r.Round.Draw(mahjong.Direction(seat))
	case "chow":
		if len(action.Tiles) < 2 {
			return errors.New("not enough tiles")
		}
		return r.Round.Chow(mahjong.Direction(seat), action.Tiles[0], action.Tiles[1])
	case "peng":
		if len(action.Tiles) < 1 {
			return errors.New("not enough tiles")
		}
		return r.Round.Peng(mahjong.Direction(seat), action.Tiles[0])
	case "kong":
		if len(action.Tiles) < 1 {
			return errors.New("not enough tiles")
		}
		return r.Round.Kong(mahjong.Direction(seat), action.Tiles[0])
	default:
		return errors.New("invalid action")
	}
}
