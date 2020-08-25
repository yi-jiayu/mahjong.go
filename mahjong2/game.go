package mahjong2

// Event represents things that happened during a mahjong game, such as
// drawing a tile, discarding a tile or creating a melded set. Events can be
// undone to return to a previous round state and vice-versa.
type Event interface {
	Undo(r *Round)
	Redo(r *Round)
}

// Tile represents a mahjong tile.
type Tile string

// MeldType represents the type of a melded set.
type MeldType string

// Allowed meld types.
const (
	MeldChi  MeldType = "chi"
	MeldPong MeldType = "pong"
	MeldGang MeldType = "gang"
	MeldEyes MeldType = "eyes"
)

// Meld represents a melded set.
type Meld struct {
	Type  MeldType
	Tiles []Tile
}

// Hand represents all the tiles belonging to a player.
type Hand struct {
	Flowers   []Tile
	Revealed  []Meld
	Concealed []Tile
}

// Player represents a participant in a mahjong game.
type Player struct {
	Name  string
	Score int
}

// Direction represents a wind direction.
type Direction int

// All possible directions.
const (
	DirectionEast Direction = iota
	DirectionSouth
	DirectionWest
	DirectionNorth
)

// Phase constrains what actions are currently possible.
type Phase string

const (
	// PhaseDraw represents the draw phase, when the player whose turn it
	// currently is may draw a tile or chi the last discarded tile, and any
	// player may pong the last discarded tile.
	PhaseDraw Phase = "draw"

	// PhaseDiscard represents the discard phase, when the player whose turn it
	// is has 14 tiles in their hand and must discard a tile. They may also
	// reveal a concealed gang or win by self-draw.
	PhaseDiscard Phase = "discard"

	// PhaseFinished represents the the end of a round, after a player wins or
	// the wall runs out of draws.
	PhaseFinished Phase = "finished"
)

// Result represents the outcome of a round.
type Result struct {
	// Dealer is the integer offset of the dealer for the round.
	Dealer int

	// Wind is the prevailing wind for the round.
	Wind Direction

	// Winner is the integer offset of the winner for the round, or -1 if the round ended in a draw.
	Winner int

	// Points is how much the winning hand was worth.
	Points int
}

// Game represents a mahjong game.
type Game struct {
	// Players contains the players for a game in order.
	Players []Player

	CurrentRound   *Round
	PreviousRounds []Round
}

// Round represents a round in a mahjong game.
type Round struct {
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
}
