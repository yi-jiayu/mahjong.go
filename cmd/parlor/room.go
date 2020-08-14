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

	// clients is a map of channels to player IDs.
	clients map[chan string]string
}

// RoomView represents a specific player or bystander's view of the room.
type RoomView struct {
	PlayerID string
	Room     *Room
}

func (r RoomView) MarshalJSON() ([]byte, error) {
	seat := -1
	for i, id := range r.Room.Players {
		if id == r.PlayerID {
			seat = i
			break
		}
	}
	players := make([]string, len(r.Room.Players))
	for i, playerID := range r.Room.Players {
		player, err := playerRepository.Get(playerID)
		if err != nil {
			players[i] = "Unknown player"
		} else {
			players[i] = player.Name
		}
	}
	v := struct {
		Seat    int                `json:"seat"`
		Nonce   int                `json:"nonce"`
		Phase   int                `json:"phase"`
		Players []string           `json:"players"`
		Round   *mahjong.RoundView `json:"round"`
	}{
		Seat:    seat,
		Nonce:   r.Room.Nonce,
		Phase:   r.Room.Phase,
		Players: players,
	}
	if r.Room.Round != nil {
		round := r.Room.Round.ViewFromSeat(mahjong.Direction(seat))
		v.Round = &round
	}
	return json.Marshal(v)
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
	r.clients = map[chan string]string{}
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

func NewRoom(host string) *Room {
	return &Room{
		Players: []string{host},
		clients: map[chan string]string{},
	}
}

func (r *Room) AddClient(playerID string, c chan string) {
	r.Lock()
	defer r.Unlock()

	r.clients[c] = playerID
	go func() {
		var b bytes.Buffer
		json.NewEncoder(&b).Encode(RoomView{
			PlayerID: playerID,
			Room:     r,
		})
		c <- b.String()
	}()
}

func (r *Room) addPlayer(id string) error {
	if len(r.Players) == 4 {
		return errors.New("room full")
	}
	for _, p := range r.Players {
		if p == id {
			return nil
		}
	}
	r.Players = append(r.Players, id)
	r.broadcast()
	return nil
}

func (r *Room) AddPlayer(id string) error {
	r.Lock()
	defer r.Unlock()
	return r.addPlayer(id)
}

func (r *Room) AddBot(playerID string, bot *Bot) error {
	r.Lock()
	defer r.Unlock()
	found := false
	for _, id := range r.Players {
		if id == playerID {
			found = true
			break
		}
	}
	if !found {
		return errors.New("not allowed")
	}
	err := r.addPlayer(bot.ID)
	if err != nil {
		return err
	}
	r.clients[bot.GameUpdates] = bot.ID
	return nil
}

func (r *Room) RemoveClient(c chan string) {
	r.Lock()
	defer r.Unlock()

	delete(r.clients, c)
}

func (r *Room) broadcast() {
	for c, playerID := range r.clients {
		var b bytes.Buffer
		json.NewEncoder(&b).Encode(RoomView{
			PlayerID: playerID,
			Room:     r,
		})
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
