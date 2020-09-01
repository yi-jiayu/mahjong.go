package parlour

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/yi-jiayu/mahjong.go"
)

type Phase int

const (
	PhaseLobby Phase = iota
	PhaseInProgress
)

var (
	errForbidden    = errors.New("forbidden")
	errInvalidNonce = errors.New("invalid nonce")
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
	Round   *mahjong.Round

	sync.RWMutex

	// clients is a map of subscription channels to player IDs.
	clients map[chan string]string
}

type RoomView struct {
	ID      string            `json:"id"`
	Nonce   int               `json:"nonce"`
	Phase   Phase             `json:"phase"`
	Players []Player          `json:"players"`
	Round   mahjong.RoundView `json:"round"`
	Inside  bool              `json:"inside"`
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
		Inside:  r.seat(playerID) != -1,
	}
	if r.Phase == PhaseInProgress {
		view.Round = r.Round.View(r.seat(playerID))
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

type ActionType string

const (
	ActionNextRound ActionType = "next"
	ActionDraw      ActionType = "draw"
	ActionDiscard   ActionType = "discard"
	ActionChi       ActionType = "chi"
	ActionPong      ActionType = "pong"
	ActionGang      ActionType = "gang"
	ActionHu        ActionType = "hu"
	ActionEndRound  ActionType = "end"
)

type Action struct {
	Nonce int            `json:"nonce"`
	Type  ActionType     `json:"type"`
	Tiles []mahjong.Tile `json:"tiles"`
}

func (r *Room) reduceRound(seat int, t time.Time, action Action) error {
	switch action.Type {
	case ActionNextRound:
		return r.nextRound()
	case ActionDraw:
		_, _, err := r.Round.Draw(seat, t)
		return err
	case ActionDiscard:
		if len(action.Tiles) < 1 {
			return errors.New("tiles is required")
		}
		return r.Round.Discard(seat, t, action.Tiles[0])
	case ActionChi:
		if len(action.Tiles) < 2 {
			return errors.New("tiles is too short")
		}
		return r.Round.Chi(seat, t, action.Tiles[0], action.Tiles[1])
	case ActionPong:
		return r.Round.Pong(seat, t)
	case ActionGang:
		if len(action.Tiles) > 0 {
			_, _, err := r.Round.GangFromHand(seat, t, action.Tiles[0])
			return err
		}
		_, _, err := r.Round.GangFromDiscard(seat, t)
		return err
	case ActionHu:
		return r.Round.Hu(seat, t)
	case ActionEndRound:
		return r.Round.End(seat, t)
	default:
		return errors.New("action is invalid")
	}
}

func (r *Room) reduce(playerID string, action Action) error {
	seat := r.seat(playerID)
	if seat == -1 {
		return errForbidden
	}
	if action.Nonce != r.Nonce {
		return errInvalidNonce
	}
	t := time.Now()
	err := r.reduceRound(seat, t, action)
	if err != nil {
		return err
	}
	r.Nonce++
	r.broadcast()
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

func (r *Room) nextRound() error {
	if r.Round == nil {
		r.Phase = PhaseInProgress
		r.Round = &mahjong.Round{
			Rules:            mahjong.RulesDefault,
			ReservedDuration: 2 * time.Second,
		}
	} else {
		next, err := r.Round.Next()
		if err != nil {
			return err
		}
		r.Round = next
	}
	r.Round.Start(rand.Int63(), time.Now())
	return nil
}

func NewRoom(host Player) *Room {
	room := &Room{
		Phase:   PhaseLobby,
		Players: []Player{host},
		clients: make(map[chan string]string),
	}
	return room
}
