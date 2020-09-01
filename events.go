package mahjong

import (
	"sort"
	"time"
)

func timeInMillis(t time.Time) int64 {
	return t.UnixNano() / 1e6
}

// EventType represents the type of an event.
type EventType string

// Possible event types.
const (
	EventStart   = "start"
	EventDraw    = "draw"
	EventDiscard = "discard"
	EventChi     = "chi"
	EventPong    = "pong"
	EventGang    = "gang"
	EventHu      = "hu"
	EventEnd     = "end"
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

// Event represents things that happened during a mahjong game, such as
// drawing a tile, discarding a tile or creating a melded set. Events can be
// undone to return to a previous round state and vice-versa.
type Event interface {
	// Undo restores a round to the state it was in before an event occurred.
	Undo(r *Round) *Round

	// Redo restores a round to the state it was in after an event occurred.
	Redo(r *Round) *Round

	// View returns a view of an Event.
	View() EventView
}

type StartEvent struct {
	Time time.Time
}

func (e StartEvent) Undo(r *Round) *Round {
	panic("implement me")
}

func (e StartEvent) Redo(r *Round) *Round {
	panic("implement me")
}

func (e StartEvent) View() EventView {
	return EventView{
		Type: EventStart,
		Time: timeInMillis(e.Time),
	}
}

type DrawEvent struct {
	Seat    int
	Time    time.Time
	Tile    Tile
	Flowers []Tile
}

func (d DrawEvent) Undo(r *Round) *Round {
	panic("implement me")
}

func (d DrawEvent) Redo(r *Round) *Round {
	panic("implement me")
}

func (d DrawEvent) View() EventView {
	return EventView{
		Type: EventDraw,
		Seat: d.Seat,
		Time: timeInMillis(d.Time),
	}
}

type DiscardEvent struct {
	Seat int
	time.Time
	Tile Tile
}

func (d DiscardEvent) Undo(r *Round) *Round {
	panic("implement me")
}

func (d DiscardEvent) Redo(r *Round) *Round {
	panic("implement me")
}

func (d DiscardEvent) View() EventView {
	return EventView{
		Type:  EventDiscard,
		Seat:  d.Seat,
		Time:  timeInMillis(d.Time),
		Tiles: []Tile{d.Tile},
	}
}

type ChiEvent struct {
	Seat        int
	Time        time.Time
	LastDiscard Tile
	Tiles       [2]Tile
}

func (c ChiEvent) Undo(r *Round) *Round {
	r.Discards = append(r.Discards, c.LastDiscard)
	hand := &r.Hands[c.Seat]
	seq := []Tile{c.LastDiscard, c.Tiles[0], c.Tiles[1]}
	sort.Slice(seq, func(i, j int) bool {
		return seq[i] < seq[j]
	})
	hand.Revealed = hand.Revealed[:len(hand.Revealed)-1]
	hand.Concealed.Add(c.Tiles[0])
	hand.Concealed.Add(c.Tiles[1])
	r.Phase = PhaseDraw
	return r
}

func (c ChiEvent) Redo(r *Round) *Round {
	r.popLastDiscard()
	hand := &r.Hands[c.Seat]
	hand.Concealed.Remove(c.Tiles[0])
	hand.Concealed.Remove(c.Tiles[1])
	seq := []Tile{c.LastDiscard, c.Tiles[0], c.Tiles[1]}
	sort.Slice(seq, func(i, j int) bool {
		return seq[i] < seq[j]
	})
	hand.Revealed = append(hand.Revealed, Meld{
		Type:  MeldChi,
		Tiles: seq,
	})
	r.Phase = PhaseDiscard
	return r
}

func (c ChiEvent) View() EventView {
	seq := []Tile{c.LastDiscard, c.Tiles[0], c.Tiles[1]}
	sort.Slice(seq, func(i, j int) bool {
		return seq[i] < seq[j]
	})
	return EventView{
		Type:  EventChi,
		Seat:  c.Seat,
		Time:  timeInMillis(c.Time),
		Tiles: seq,
	}
}

type PongEvent struct {
	Seat         int
	Time         time.Time
	Tile         Tile
	PreviousTurn int
}

func (p PongEvent) Undo(r *Round) *Round {
	panic("implement me")
}

func (p PongEvent) Redo(r *Round) *Round {
	panic("implement me")
}

func (p PongEvent) View() EventView {
	return EventView{
		Type:  EventPong,
		Seat:  p.Seat,
		Time:  timeInMillis(p.Time),
		Tiles: []Tile{p.Tile},
	}
}

type GangEvent struct {
	Seat         int
	Time         time.Time
	Tile         Tile
	Replacement  Tile
	Flowers      []Tile
	PreviousTurn int
}

func (g GangEvent) Undo(r *Round) *Round {
	panic("implement me")
}

func (g GangEvent) Redo(r *Round) *Round {
	panic("implement me")
}

func (g GangEvent) View() EventView {
	return EventView{
		Type:  EventGang,
		Seat:  g.Seat,
		Time:  timeInMillis(g.Time),
		Tiles: append([]Tile{g.Tile}, g.Flowers...),
	}
}

type HuEvent struct {
	Seat int
	Time time.Time
}

func (h HuEvent) Undo(r *Round) *Round {
	panic("implement me")
}

func (h HuEvent) Redo(r *Round) *Round {
	panic("implement me")
}

func (h HuEvent) View() EventView {
	return EventView{
		Type: EventHu,
		Seat: h.Seat,
		Time: timeInMillis(h.Time),
	}
}

type EndEvent struct {
	Seat int
	Time time.Time
}

func (e EndEvent) Undo(r *Round) *Round {
	panic("implement me")
}

func (e EndEvent) Redo(r *Round) *Round {
	panic("implement me")
}

func (e EndEvent) View() EventView {
	return EventView{
		Type: EventEnd,
		Seat: e.Seat,
		Time: timeInMillis(e.Time),
	}
}
