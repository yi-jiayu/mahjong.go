package mahjong2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testEvent(t *testing.T, before, after func() *Round, event Event) {
	t.Run("undo", func(t *testing.T) {
		assert.Equal(t, before(), event.Undo(after()))
	})
	t.Run("redo", func(t *testing.T) {
		assert.Equal(t, after(), event.Redo(before()))
	})
}

func TestChiEvent_UndoRedo(t *testing.T) {
	before := func() *Round {
		return &Round{
			Turn:     0,
			Phase:    PhaseDraw,
			Discards: []Tile{TileBamboo4, TileBamboo3},
			Hands: []Hand{{
				Revealed:  []Meld{},
				Concealed: NewTileBag([]Tile{TileWindsWest, TileBamboo1, TileBamboo2}),
			}},
		}
	}
	after := func() *Round {
		return &Round{
			Turn:     0,
			Phase:    PhaseDiscard,
			Discards: []Tile{TileBamboo4},
			Hands: []Hand{{
				Revealed: []Meld{{
					Type:  MeldChi,
					Tiles: []Tile{TileBamboo1, TileBamboo2, TileBamboo3},
				}},
				Concealed: NewTileBag([]Tile{TileWindsWest}),
			}},
		}
	}
	chi := ChiEvent{
		Seat:        0,
		LastDiscard: TileBamboo3,
		Tiles:       [2]Tile{TileBamboo1, TileBamboo2},
	}
	testEvent(t, before, after, chi)
}
