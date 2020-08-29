package mahjong

import (
	"time"
)

// EventType represents the type of an event.
type EventType string

// Possible event types.
const (
	EventDraw    = "draw"
	EventDiscard = "discard"
	EventChi     = "chi"
	EventPong    = "pong"
	EventGang    = "gang"
)

// EventView represents a player's view of an event.
type EventView struct {
	// Type is the type of an event.
	Type EventType `json:"type"`

	// Seat is the integer offset of the player an event pertains to.
	Seat int `json:"seat"`

	// Time is the time an event occurred.
	Time time.Time `json:"time"`

	// Tiles are the tiles involved in an event.
	Tiles []Tile `json:"tiles"`
}

// RoundView represents a player's view of a round.
type RoundView struct {
	Seat             int           `json:"seat"`
	Scores           []int         `json:"scores"`
	Hands            []Hand        `json:"hands"`
	DrawsLeft        int           `json:"draws_left"`
	Discards         []Tile        `json:"discards"`
	Wind             Direction     `json:"wind"`
	Dealer           int           `json:"dealer"`
	Turn             int           `json:"turn"`
	Phase            Phase         `json:"phase"`
	Events           []EventView   `json:"events"`
	Result           Result        `json:"result"`
	LastDiscardTime  time.Time     `json:"last_discard_time"`
	ReservedDuration time.Duration `json:"reserved_duration"`
}

// GameView represents a player's view of the game.
type GameView struct {
	CurrentRound RoundView `json:"current_round"`
	Results      []Result  `json:"results"`
}
