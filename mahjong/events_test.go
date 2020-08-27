package mahjong

import (
	"testing"
	"time"

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

func TestDrawEvent_View(t *testing.T) {
	now := time.Now()
	draw := DrawEvent{
		Seat: 1,
		Time: now,
	}
	assert.Equal(t, EventView{
		Type: EventDraw,
		Seat: 1,
		Time: now,
	}, draw.View())
}

func TestDiscardEvent_View(t *testing.T) {
	now := time.Now()
	draw := DiscardEvent{
		Seat: 1,
		Time: now,
		Tile: TileWindsWest,
	}
	assert.Equal(t, EventView{
		Type:  EventDiscard,
		Seat:  1,
		Time:  now,
		Tiles: []Tile{TileWindsWest},
	}, draw.View())
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

func TestChiEvent_View(t *testing.T) {
	now := time.Now()
	chi := ChiEvent{
		Seat:        1,
		Time:        now,
		LastDiscard: TileDots4,
		Tiles:       [2]Tile{TileDots5, TileDots3},
	}
	assert.Equal(t, EventView{
		Type:  EventChi,
		Seat:  1,
		Time:  now,
		Tiles: []Tile{TileDots3, TileDots4, TileDots5},
	}, chi.View())
}

func TestPongEvent_View(t *testing.T) {
	now := time.Now()
	pong := PongEvent{
		Seat: 1,
		Time: now,
		Tile: TileDragonsRed,
	}
	assert.Equal(t, EventView{
		Type:  EventPong,
		Seat:  1,
		Time:  now,
		Tiles: []Tile{TileDragonsRed},
	}, pong.View())
}

func TestGangEvent_View(t *testing.T) {
	now := time.Now()
	gang := GangEvent{
		Seat:    1,
		Time:    now,
		Tile:    TileDragonsGreen,
		Flowers: []Tile{TileRooster, TileCentipede},
	}
	assert.Equal(t, EventView{
		Type:  EventGang,
		Seat:  1,
		Time:  now,
		Tiles: []Tile{TileDragonsGreen, TileRooster, TileCentipede},
	}, gang.View())
}
