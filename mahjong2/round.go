package mahjong2

import (
	"time"
)

// Round represents a round in a mahjong game.
type Round struct {
	// Scores contains the score for each player in the game.
	Scores []int

	// Hands contains the corresponding hand for each player in the game.
	Hands []Hand

	// Wall contains the remaining tiles left to be drawn.
	Wall []Tile

	// Discards contains all the previously discarded tiles.
	Discards []Tile

	// Wind is the prevailing wind for the round.
	Wind Direction

	// Dealer is the integer offset of the dealer for the round.
	Dealer int

	// Turn is the integer offset of the player whose turn it currently is.
	Turn int

	// Phase is the current turn phase.
	Phase Phase

	// Events contains all the events that happened in the round.
	Events []Event

	// Result is the outcome of the round.
	Result Result

	LastDiscardTime  time.Time
	ReservedDuration time.Duration
}

func (r *Round) lastDiscard() Tile {
	if len(r.Discards) == 0 {
		return ""
	}
	return r.Discards[len(r.Discards)-1]
}

func (r *Round) popLastDiscard() Tile {
	tile := r.Discards[len(r.Discards)-1]
	r.Discards = r.Discards[:len(r.Discards)-1]
	return tile
}

func (r *Round) drawFront() Tile {
	drawn := r.Wall[0]
	r.Wall = r.Wall[1:]
	return drawn
}

func (r *Round) drawBack() Tile {
	drawn := r.Wall[len(r.Wall)-1]
	r.Wall = r.Wall[:len(r.Wall)-1]
	return drawn
}

func (r *Round) previousTurn() int {
	return (r.Turn + 3) % 4
}

func (r *Round) replaceTile() (Tile, []Tile) {
	flowers := make([]Tile, 0)
	drawn := r.drawBack()
	for isFlower(drawn) {
		flowers = append(flowers, drawn)
		drawn = r.drawBack()
	}
	return drawn, flowers
}
