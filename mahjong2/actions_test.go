package mahjong2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRound_Draw(t *testing.T) {
	t.Run("cannot draw on wrong turn", func(t *testing.T) {
		r := &Round{Turn: 0}
		_, _, err := r.Draw(1, time.Now())
		assert.EqualError(t, err, "wrong turn")
	})
	t.Run("can only draw during draw phase", func(t *testing.T) {
		r := &Round{Turn: 0, Phase: PhaseDiscard}
		_, _, err := r.Draw(0, time.Now())
		assert.EqualError(t, err, "wrong phase")
	})
	t.Run("cannot draw during reserved duration", func(t *testing.T) {
		now := time.Now()
		oneSecondAgo := now.Add(-time.Second)
		r := &Round{
			Turn:             0,
			Phase:            PhaseDraw,
			LastDiscardTime:  oneSecondAgo,
			ReservedDuration: 2 * time.Second,
		}
		_, _, err := r.Draw(0, time.Now())
		assert.EqualError(t, err, "cannot draw during reserved duration")
	})
	t.Run("successful draw", func(t *testing.T) {
		seat := 0
		r := &Round{
			Wall:  []Tile{TileBamboo1, TileDots5},
			Turn:  seat,
			Phase: PhaseDraw,
			Hands: []Hand{{Concealed: NewTileBag([]Tile{TileWindsWest})}},
		}
		now := time.Now()
		drawn, flowers, err := r.Draw(seat, now)
		assert.NoError(t, err)
		assert.Equal(t, TileBamboo1, drawn)
		assert.Empty(t, flowers)
		assert.Equal(t, []Tile{TileDots5}, r.Wall)
		assert.Equal(t, NewTileBag([]Tile{TileWindsWest, TileBamboo1}), r.Hands[seat].Concealed)
		assert.Equal(t, seat, r.Turn)
		assert.Equal(t, PhaseDiscard, r.Phase)
	})
	t.Run("successful draw with flowers", func(t *testing.T) {
		seat := 0
		r := &Round{
			Wall:  []Tile{TileGentlemen1, TileBamboo1, TileDots5, TileGentlemen2},
			Turn:  seat,
			Phase: PhaseDraw,
			Hands: []Hand{{Concealed: NewTileBag([]Tile{TileWindsWest})}},
		}
		now := time.Now()
		drawn, flowers, err := r.Draw(seat, now)
		assert.NoError(t, err)
		assert.Equal(t, TileDots5, drawn)
		assert.Equal(t, []Tile{TileGentlemen1, TileGentlemen2}, flowers)
		assert.Equal(t, []Tile{TileBamboo1}, r.Wall)
		assert.Equal(t, NewTileBag([]Tile{TileWindsWest, TileDots5}), r.Hands[seat].Concealed)
		assert.Equal(t, seat, r.Turn)
		assert.Equal(t, PhaseDiscard, r.Phase)
	})
}

func TestRound_Chi(t *testing.T) {
	t.Run("cannot chi on wrong turn", func(t *testing.T) {
		r := &Round{Turn: 0}
		err := r.Chi(1, time.Now(), "", "")
		assert.EqualError(t, err, "wrong turn")
	})
	t.Run("can only chi during draw phase", func(t *testing.T) {
		r := &Round{Turn: 0, Phase: PhaseDiscard}
		err := r.Chi(0, time.Now(), "", "")
		assert.EqualError(t, err, "wrong phase")
	})
	t.Run("cannot chi when no discards", func(t *testing.T) {
		r := &Round{
			Turn:  0,
			Phase: PhaseDraw,
		}
		err := r.Chi(0, time.Now(), TileBamboo1, TileBamboo2)
		assert.EqualError(t, err, "no discards")
	})
	t.Run("cannot chi non-suited tile", func(t *testing.T) {
		r := &Round{
			Turn:     0,
			Phase:    PhaseDraw,
			Discards: []Tile{TileDragonsRed},
		}
		err := r.Chi(0, time.Now(), "", "")
		assert.EqualError(t, err, "cannot chi non-suited tile")
	})
	t.Run("cannot chi with invalid sequence", func(t *testing.T) {
		r := &Round{
			Turn:     0,
			Phase:    PhaseDraw,
			Discards: []Tile{TileBamboo3},
		}
		err := r.Chi(0, time.Now(), TileCharacters2, TileCharacters4)
		assert.EqualError(t, err, "invalid sequence")
	})
	t.Run("cannot chi when missing tiles", func(t *testing.T) {
		r := &Round{
			Turn:     0,
			Phase:    PhaseDraw,
			Discards: []Tile{TileBamboo3},
			Hands:    []Hand{{}},
		}
		err := r.Chi(0, time.Now(), TileBamboo2, TileBamboo4)
		assert.EqualError(t, err, "missing tiles")
	})
	t.Run("cannot chi during reserved duration", func(t *testing.T) {
		now := time.Now()
		oneSecondAgo := now.Add(-time.Second)
		r := &Round{
			Turn:             0,
			Phase:            PhaseDraw,
			Discards:         []Tile{TileBamboo4, TileBamboo3},
			Hands:            []Hand{{Concealed: NewTileBag([]Tile{TileWindsWest, TileBamboo1, TileBamboo2})}},
			LastDiscardTime:  oneSecondAgo,
			ReservedDuration: 2 * time.Second,
		}
		err := r.Chi(0, time.Now(), TileBamboo1, TileBamboo2)
		assert.EqualError(t, err, "cannot chi during reserved duration")
	})
	t.Run("successful chi", func(t *testing.T) {
		now := time.Now()
		r := &Round{
			Turn:     0,
			Phase:    PhaseDraw,
			Discards: []Tile{TileBamboo4, TileBamboo3},
			Hands:    []Hand{{Concealed: NewTileBag([]Tile{TileWindsWest, TileBamboo1, TileBamboo2})}},
		}
		err := r.Chi(0, now, TileBamboo1, TileBamboo2)
		assert.NoError(t, err)
		assert.Equal(t, []Tile{TileBamboo4}, r.Discards)
		assert.Equal(t, NewTileBag([]Tile{TileWindsWest}), r.Hands[0].Concealed)
		assert.Equal(t, []Meld{{
			Type:  MeldChi,
			Tiles: []Tile{TileBamboo1, TileBamboo2, TileBamboo3},
		}}, r.Hands[0].Revealed)
		assert.Equal(t, 0, r.Turn)
		assert.Equal(t, PhaseDiscard, r.Phase)
	})
}

func TestRound_Pong(t *testing.T) {
	t.Run("cannot pong immediately after own turn", func(t *testing.T) {
		r := &Round{Turn: 1}
		err := r.Pong(0, time.Now())
		assert.EqualError(t, err, "wrong turn")
	})
	t.Run("can only pong during draw phase", func(t *testing.T) {
		r := &Round{Turn: 0, Phase: PhaseDiscard}
		err := r.Pong(0, time.Now())
		assert.EqualError(t, err, "wrong phase")
	})
	t.Run("cannot pong when no discards", func(t *testing.T) {
		r := &Round{
			Turn:  0,
			Phase: PhaseDraw,
		}
		err := r.Pong(0, time.Now())
		assert.EqualError(t, err, "no discards")
	})
	t.Run("cannot pong when missing tiles", func(t *testing.T) {
		r := &Round{
			Turn:     0,
			Phase:    PhaseDraw,
			Discards: []Tile{TileDragonsRed},
			Hands:    []Hand{{}},
		}
		err := r.Pong(0, time.Now())
		assert.EqualError(t, err, "missing tiles")
	})
	t.Run("successful pong", func(t *testing.T) {
		seat := 2
		r := &Round{
			Turn:     seat,
			Phase:    PhaseDraw,
			Discards: []Tile{TileDots1, TileDragonsRed},
			Hands:    []Hand{{}, {}, {Concealed: NewTileBag([]Tile{TileWindsWest, TileDragonsRed, TileDragonsRed})}},
		}
		err := r.Pong(seat, time.Now())
		assert.NoError(t, err)
		assert.Equal(t, []Tile{TileDots1}, r.Discards)
		assert.Equal(t, NewTileBag([]Tile{TileWindsWest}), r.Hands[seat].Concealed)
		assert.Equal(t, []Meld{{
			Type:  MeldPong,
			Tiles: []Tile{TileDragonsRed},
		}}, r.Hands[seat].Revealed)
		assert.Equal(t, seat, r.Turn)
		assert.Equal(t, PhaseDiscard, r.Phase)
	})
}

func TestRound_GangFromDiscard(t *testing.T) {
	t.Run("cannot gang from discard immediately after own turn", func(t *testing.T) {
		r := &Round{Turn: 1}
		_, _, err := r.GangFromDiscard(0, time.Now())
		assert.EqualError(t, err, "wrong turn")
	})
	t.Run("can only gang from discard during draw phase", func(t *testing.T) {
		r := &Round{Turn: 0, Phase: PhaseDiscard}
		_, _, err := r.GangFromDiscard(0, time.Now())
		assert.EqualError(t, err, "wrong phase")
	})
	t.Run("cannot gang from discard when no discards", func(t *testing.T) {
		r := &Round{
			Turn:  0,
			Phase: PhaseDraw,
		}
		_, _, err := r.GangFromDiscard(0, time.Now())
		assert.EqualError(t, err, "no discards")
	})
	t.Run("cannot gang from discard when not enough tiles", func(t *testing.T) {
		r := &Round{
			Turn:     0,
			Phase:    PhaseDraw,
			Discards: []Tile{TileDragonsRed},
			Hands:    []Hand{{Concealed: TileBag{TileDragonsRed: 2}}},
		}
		_, _, err := r.GangFromDiscard(0, time.Now())
		assert.EqualError(t, err, "missing tiles")
	})
	t.Run("successful gang from discard", func(t *testing.T) {
		seat := 2
		r := &Round{
			Wall:     []Tile{TileCharacters4, TileCharacters6, TileGentlemen1},
			Turn:     seat,
			Phase:    PhaseDraw,
			Discards: []Tile{TileDots1, TileDragonsRed},
			Hands: []Hand{{}, {}, {
				Flowers:   []Tile{TileCat},
				Concealed: NewTileBag([]Tile{TileWindsWest, TileDragonsRed, TileDragonsRed, TileDragonsRed}),
			}},
		}
		replacement, flowers, err := r.GangFromDiscard(seat, time.Now())
		assert.NoError(t, err)
		assert.Equal(t, TileCharacters6, replacement)
		assert.Equal(t, []Tile{TileGentlemen1}, flowers)
		assert.Equal(t, []Tile{TileDots1}, r.Discards)
		assert.Equal(t, []Tile{TileCat, TileGentlemen1}, r.Hands[seat].Flowers)
		assert.Equal(t, NewTileBag([]Tile{TileWindsWest, TileCharacters6}), r.Hands[seat].Concealed)
		assert.Equal(t, []Meld{{
			Type:  MeldGang,
			Tiles: []Tile{TileDragonsRed},
		}}, r.Hands[seat].Revealed)
		assert.Equal(t, seat, r.Turn)
		assert.Equal(t, PhaseDiscard, r.Phase)
	})
}

func TestRound_GangFromHand(t *testing.T) {
	t.Run("cannot gang from hand on wrong turn", func(t *testing.T) {
		r := &Round{Turn: 1}
		_, _, err := r.GangFromHand(0, time.Now(), "")
		assert.EqualError(t, err, "wrong turn")
	})
	t.Run("can only gang from hand during discard phase", func(t *testing.T) {
		r := &Round{Turn: 0, Phase: PhaseDraw}
		_, _, err := r.GangFromHand(0, time.Now(), "")
		assert.EqualError(t, err, "wrong phase")
	})
	t.Run("cannot gang from hand when not enough tiles and no corresponding pong", func(t *testing.T) {
		r := &Round{
			Turn:  0,
			Phase: PhaseDiscard,
			Hands: []Hand{{}},
		}
		_, _, err := r.GangFromHand(0, time.Now(), TileDragonsRed)
		assert.EqualError(t, err, "missing tiles")
	})
	t.Run("successful concealed gang", func(t *testing.T) {
		seat := 0
		r := &Round{
			Wall:  []Tile{TileCharacters1, TileDots4, TileCat},
			Turn:  0,
			Phase: PhaseDiscard,
			Hands: []Hand{{
				Flowers:   []Tile{TileSeasons1},
				Concealed: TileBag{TileDragonsRed: 4},
			}},
		}
		replacement, flowers, err := r.GangFromHand(seat, time.Now(), TileDragonsRed)
		assert.NoError(t, err)
		assert.Equal(t, TileDots4, replacement)
		assert.Equal(t, []Tile{TileCat}, flowers)
		assert.Equal(t, []Meld{{
			Type:  MeldGang,
			Tiles: []Tile{TileDragonsRed},
		}}, r.Hands[seat].Revealed)
		assert.Equal(t, TileBag{TileDots4: 1}, r.Hands[seat].Concealed)
		assert.Equal(t, []Tile{TileCharacters1}, r.Wall)
		assert.Equal(t, seat, r.Turn)
		assert.Equal(t, PhaseDiscard, r.Phase)
	})
	t.Run("successful promote pong to gang", func(t *testing.T) {
		seat := 0
		r := &Round{
			Wall:  []Tile{TileCharacters1, TileDots4, TileCat},
			Turn:  0,
			Phase: PhaseDiscard,
			Hands: []Hand{{
				Flowers: []Tile{TileSeasons1},
				Revealed: []Meld{{
					Type:  MeldPong,
					Tiles: []Tile{TileDragonsRed},
				}},
				Concealed: TileBag{TileDragonsRed: 1},
			}},
		}
		replacement, flowers, err := r.GangFromHand(seat, time.Now(), TileDragonsRed)
		assert.NoError(t, err)
		assert.Equal(t, TileDots4, replacement)
		assert.Equal(t, []Tile{TileCat}, flowers)
		assert.Equal(t, []Meld{{
			Type:  MeldGang,
			Tiles: []Tile{TileDragonsRed},
		}}, r.Hands[seat].Revealed)
		assert.Equal(t, TileBag{TileDots4: 1}, r.Hands[seat].Concealed)
		assert.Equal(t, []Tile{TileCharacters1}, r.Wall)
		assert.Equal(t, seat, r.Turn)
		assert.Equal(t, PhaseDiscard, r.Phase)
	})
}
