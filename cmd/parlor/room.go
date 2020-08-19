package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/yi-jiayu/mahjong.go"
)

const (
	PhaseLobby = iota
	PhaseInProgress
	PhaseRoundOver
)

// Result represents the result of a round.
type Result struct {
	// Dealer is the integer offset of the dealer for a particular round.
	Dealer int `json:"dealer"`

	// PrevailingWind was the prevailing wind for a particular round.
	PrevailingWind mahjong.Direction `json:"prevailing_wind"`

	// Winner is the integer offset of the winner for a particular round.
	Winner int `json:"winner"`

	Points int `json:"points"`
}

type Room struct {
	ID      string
	Nonce   int
	Phase   int
	Players []string
	Round   *mahjong.Round

	// CurrentDealer is the integer offset of the current dealer in the players array.
	CurrentDealer int

	// PrevailingWind is the prevailing wind for the current round.
	PrevailingWind mahjong.Direction

	// Results contains the results of previous rounds played in this room.
	Results []Result

	sync.RWMutex

	// clients is a map of channels to player IDs.
	clients map[chan string]string
}

// RoomView represents a specific player or bystander's view of the room.
type RoomView struct {
	Seat           mahjong.Direction  `json:"seat"`
	Nonce          int                `json:"nonce"`
	Phase          int                `json:"phase"`
	Players        []string           `json:"players"`
	Results        []Result           `json:"results"`
	PrevailingWind mahjong.Direction  `json:"prevailing_wind"`
	Round          *mahjong.RoundView `json:"round"`
}

func (r *Room) seat(playerID string) int {
	for i, id := range r.Players {
		if id == playerID {
			return (i - r.CurrentDealer + 4) % 4
		}
	}
	return -1
}

func (r *Room) ViewAs(playerID string) RoomView {
	seat := r.seat(playerID)
	players := make([]string, len(r.Players))
	for i, playerID := range r.Players {
		player, err := playerRepository.Get(playerID)
		if err != nil {
			players[(i-r.CurrentDealer+4)%4] = "Unknown player"
		} else {
			players[(i-r.CurrentDealer+4)%4] = player.Name
		}
	}
	view := RoomView{
		Seat:           mahjong.Direction(seat),
		Nonce:          r.Nonce,
		Phase:          r.Phase,
		Players:        players,
		Results:        r.Results,
		PrevailingWind: r.PrevailingWind,
	}
	if r.Round != nil {
		round := r.Round.ViewFromSeat(mahjong.Direction(seat))
		view.Round = &round
	}
	return view
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
		json.NewEncoder(&b).Encode(r.ViewAs(playerID))
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
		json.NewEncoder(&b).Encode(r.ViewAs(playerID))
		c <- b.String()
	}
}

func (r *Room) startRound() error {
	if len(r.Players) < 4 {
		return errors.New("not enough players")
	}
	r.Round = mahjong.NewRound(rand.Int63())
	r.Round.PongDuration = 2 * time.Second
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
	seat := r.seat(playerID)
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
		return nil, r.Round.Discard(mahjong.Direction(seat), action.Tiles[0], time.Now())
	case "draw":
		drawn, flowers, err := r.Round.Draw(mahjong.Direction(seat), time.Now())
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
		return nil, r.Round.Chow(mahjong.Direction(seat), action.Tiles[0], action.Tiles[1], time.Now())
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
		result := Result{
			Dealer:         r.CurrentDealer,
			PrevailingWind: r.PrevailingWind,
			Winner:         (r.CurrentDealer + int(r.Round.CurrentTurn)) % 4,
		}
		r.Results = append(r.Results, result)
		return nil, nil
	case "next round":
		if r.Phase != PhaseRoundOver {
			return nil, errors.New("the current round is not over")
		}
		if r.Round.CurrentTurn != 0 {
			if r.CurrentDealer == 3 {
				r.PrevailingWind++
			}
			r.CurrentDealer = (r.CurrentDealer + 1) % 4
		}
		r.Round = mahjong.NewRound(rand.Int63())
		r.Round.PongDuration = 2 * time.Second
		r.Phase = PhaseInProgress
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

func DispatchAction(roomID, playerID string, action Action) (interface{}, error) {
	room, err := roomRepository.Get(roomID)
	if err != nil {
		return nil, err
	}
	return room.HandleAction(playerID, action)
}
