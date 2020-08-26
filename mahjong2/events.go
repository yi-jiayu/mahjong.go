package mahjong2

import (
	"sort"
)

// Event represents things that happened during a mahjong game, such as
// drawing a tile, discarding a tile or creating a melded set. Events can be
// undone to return to a previous round state and vice-versa.
type Event interface {
	Undo(r *Round) *Round
	Redo(r *Round) *Round
}

type ChiEvent struct {
	Seat        int
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
