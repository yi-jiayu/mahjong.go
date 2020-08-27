package mahjong

import (
	"sort"
	"time"
)

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
		Time: d.Time,
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
		Time:  d.Time,
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
		Time:  c.Time,
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
		Time:  p.Time,
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
		Time:  g.Time,
		Tiles: append([]Tile{g.Tile}, g.Flowers...),
	}
}
