package mahjong2

import (
	"time"
)

// HandView represents a player's view of another player's hand.
type HandView struct {
	Flowers  []Tile
	Revealed []Meld

	// Concealed is how many concealed tiles the player has.
	Concealed int
}

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
	Type EventType

	// Seat is the integer offset of the player an event pertains to.
	Seat int

	// Time is the time an event occurred.
	Time time.Time

	// Tiles are the tiles involved in an event.
	Tiles []Tile
}

// RoundView represents a player's view of a round.
type RoundView struct {
	Seat      int
	Hand      Hand
	Opponents HandView
	DrawsLeft int
	Discards  []Tile
	Wind      Direction
	Dealer    int
	Turn      int
	Phase     Phase
	Events    []EventView
}

// GameView represents a player's view of the game.
type GameView struct {
	CurrentRound         RoundView
	PreviousRoundResults []Result
}
