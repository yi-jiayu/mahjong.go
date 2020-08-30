package mahjong

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
	Time int64 `json:"time"`

	// Tiles are the tiles involved in an event.
	Tiles []Tile `json:"tiles"`
}

// RoundView represents a player's view of a round.
type RoundView struct {
	Seat      int         `json:"seat"`
	Scores    []int       `json:"scores"`
	Hands     []Hand      `json:"hands"`
	DrawsLeft int         `json:"draws_left"`
	Discards  []Tile      `json:"discards"`
	Wind      Direction   `json:"wind"`
	Dealer    int         `json:"dealer"`
	Turn      int         `json:"turn"`
	Phase     Phase       `json:"phase"`
	Events    []EventView `json:"events"`
	Result    Result      `json:"result"`

	// LastActionTime is the time the last action took place represented in milliseconds since the Unix epoch.
	LastActionTime int64 `json:"last_action_time"`

	// ReservedDuration is a duration in milliseconds reserved for players to pong or gang after a discard.
	ReservedDuration int64 `json:"reserved_duration"`
}
