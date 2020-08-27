package mahjong

import (
	"math/rand"
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
		assert.Equal(t, []Event{DrawEvent{
			Seat:    seat,
			Time:    now,
			Tile:    drawn,
			Flowers: flowers,
		}}, r.Events)
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
		assert.Equal(t, []Event{DrawEvent{
			Seat:    seat,
			Time:    now,
			Tile:    drawn,
			Flowers: flowers,
		}}, r.Events)
	})
}

func TestRound_Discard(t *testing.T) {
	t.Run("can only discard during own turn", func(t *testing.T) {
		r := &Round{Turn: 0}
		err := r.Discard(1, time.Now(), "")
		assert.EqualError(t, err, "wrong turn")
	})
	t.Run("can only discard during discard phase", func(t *testing.T) {
		r := &Round{Turn: 0, Phase: PhaseDraw}
		err := r.Discard(0, time.Now(), "")
		assert.EqualError(t, err, "wrong phase")
	})
	t.Run("cannot discard when missing tile", func(t *testing.T) {
		r := &Round{
			Turn:  0,
			Phase: PhaseDiscard,
			Hands: []Hand{{}},
		}
		err := r.Discard(0, time.Now(), TileDragonsRed)
		assert.EqualError(t, err, "missing tiles")
	})
	t.Run("successful discard", func(t *testing.T) {
		seat := 0
		r := &Round{
			Turn:     seat,
			Phase:    PhaseDiscard,
			Discards: []Tile{TileWindsEast},
			Hands:    []Hand{{Concealed: NewTileBag([]Tile{TileCharacters1, TileWindsNorth})}},
		}
		now := time.Now()
		err := r.Discard(seat, now, TileWindsNorth)
		assert.NoError(t, err)
		assert.Equal(t, NewTileBag([]Tile{TileCharacters1}), r.Hands[seat].Concealed)
		assert.Equal(t, []Tile{TileWindsEast, TileWindsNorth}, r.Discards)
		assert.Equal(t, 1, r.Turn)
		assert.Equal(t, PhaseDraw, r.Phase)
		assert.Equal(t, []Event{DiscardEvent{
			Seat: seat,
			Time: now,
			Tile: TileWindsNorth,
		}}, r.Events)
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
		assert.EqualError(t, err, "invalid sequence")
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
		assert.Equal(t, []Event{ChiEvent{
			Seat:        0,
			Time:        now,
			LastDiscard: TileBamboo3,
			Tiles:       [2]Tile{TileBamboo1, TileBamboo2},
		}}, r.Events)
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
		seat := 1
		r := &Round{
			Turn:     3,
			Phase:    PhaseDraw,
			Discards: []Tile{TileDots1, TileDragonsRed},
			Hands:    []Hand{{}, {Concealed: NewTileBag([]Tile{TileWindsWest, TileDragonsRed, TileDragonsRed})}},
		}
		now := time.Now()
		err := r.Pong(seat, now)
		assert.NoError(t, err)
		assert.Equal(t, []Tile{TileDots1}, r.Discards)
		assert.Equal(t, NewTileBag([]Tile{TileWindsWest}), r.Hands[seat].Concealed)
		assert.Equal(t, []Meld{{
			Type:  MeldPong,
			Tiles: []Tile{TileDragonsRed},
		}}, r.Hands[seat].Revealed)
		assert.Equal(t, seat, r.Turn)
		assert.Equal(t, PhaseDiscard, r.Phase)
		assert.Equal(t, []Event{PongEvent{
			Seat:         seat,
			Time:         now,
			Tile:         TileDragonsRed,
			PreviousTurn: 3,
		}}, r.Events)
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
		seat := 1
		r := &Round{
			Wall:     []Tile{TileCharacters4, TileCharacters6, TileGentlemen1},
			Turn:     3,
			Phase:    PhaseDraw,
			Discards: []Tile{TileDots1, TileDragonsRed},
			Hands: []Hand{{}, {
				Flowers:   []Tile{TileCat},
				Concealed: NewTileBag([]Tile{TileWindsWest, TileDragonsRed, TileDragonsRed, TileDragonsRed}),
			}},
		}
		now := time.Now()
		replacement, flowers, err := r.GangFromDiscard(seat, now)
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
		assert.Equal(t, []Event{GangEvent{
			Seat: seat,
			Time: now,
			Tile: TileDragonsRed,
		}}, r.Events)
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
		now := time.Now()
		replacement, flowers, err := r.GangFromHand(seat, now, TileDragonsRed)
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
		assert.Equal(t, []Event{GangEvent{
			Seat: seat,
			Time: now,
			Tile: TileDragonsRed,
		}}, r.Events)
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
		now := time.Now()
		replacement, flowers, err := r.GangFromHand(seat, now, TileDragonsRed)
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
		assert.Equal(t, []Event{GangEvent{
			Seat: seat,
			Time: now,
			Tile: TileDragonsRed,
		}}, r.Events)
	})
}

func TestRound_Hu(t *testing.T) {
	t.Run("cannot hu immediately after discarding", func(t *testing.T) {
		r := &Round{Turn: 1}
		err := r.Hu(0, time.Now())
		assert.EqualError(t, err, "wrong turn")
	})
	t.Run("cannot hu during own draw phase", func(t *testing.T) {
		r := &Round{Turn: 0, Phase: PhaseDraw}
		err := r.Hu(0, time.Now())
		assert.EqualError(t, err, "wrong phase")
	})
	t.Run("cannot hu during other player's discard phase", func(t *testing.T) {
		r := &Round{Turn: 2, Phase: PhaseDiscard}
		err := r.Hu(0, time.Now())
		assert.EqualError(t, err, "wrong turn")
	})
	t.Run("cannot hu when winning hand cannot be found", func(t *testing.T) {
		r := &Round{
			Turn:  0,
			Phase: PhaseDiscard,
			Hands: []Hand{{Concealed: TileBag{}}},
		}
		err := r.Hu(0, time.Now())
		assert.EqualError(t, err, "missing tiles")
	})
	t.Run("successful zi mo hu", func(t *testing.T) {
		seat := 1
		r := &Round{
			Turn:  seat,
			Phase: PhaseDiscard,
			Hands: []Hand{{},
				{
					Flowers:  []Tile{TileGentlemen1, TileCat},
					Revealed: []Meld{{Type: MeldChi, Tiles: []Tile{TileDots3, TileDots4, TileDots5}}},
					Concealed: NewTileBag([]Tile{
						TileBamboo6, TileBamboo7, TileBamboo8,
						TileWindsWest, TileWindsWest, TileWindsWest,
						TileCharacters8, TileCharacters8, TileCharacters8,
						TileDragonsWhite, TileDragonsWhite,
					}),
				},
			},
		}
		err := r.Hu(seat, time.Now())
		assert.NoError(t, err)
		assert.Equal(t, PhaseFinished, r.Phase)
		assert.Equal(t, Result{
			Winner: seat,
			WinningTiles: []Tile{
				"05梅", "01猫",
				"15三筒", "16四筒", "17五筒",
				"27六索", "28七索", "29八索",
				"38八万", "38八万", "38八万",
				"42西风", "42西风", "42西风",
				"46白板", "46白板",
			},
		}, r.Result)
	})
	t.Run("successful hu from discards", func(t *testing.T) {
		seat := 2
		r := &Round{
			Turn:     0,
			Phase:    PhaseDraw,
			Discards: []Tile{TileDragonsRed, TileDragonsWhite},
			Hands: []Hand{{}, {},
				{
					Flowers:  []Tile{TileGentlemen1, TileCat},
					Revealed: []Meld{{Type: MeldChi, Tiles: []Tile{TileDots3, TileDots4, TileDots5}}},
					Concealed: NewTileBag([]Tile{
						TileBamboo6, TileBamboo7, TileBamboo8,
						TileWindsWest, TileWindsWest, TileWindsWest,
						TileCharacters8, TileCharacters8, TileCharacters8,
						TileDragonsWhite,
					}),
				},
			},
		}
		err := r.Hu(seat, time.Now())
		assert.NoError(t, err)
		assert.Equal(t, []Tile{TileDragonsRed}, r.Discards)
		assert.Equal(t, PhaseFinished, r.Phase)
		assert.Equal(t, Result{
			Winner: seat,
			WinningTiles: []Tile{
				"05梅", "01猫",
				"15三筒", "16四筒", "17五筒",
				"27六索", "28七索", "29八索",
				"38八万", "38八万", "38八万",
				"42西风", "42西风", "42西风",
				"46白板", "46白板",
			},
		}, r.Result)
	})
}

func TestRound_View(t *testing.T) {
	r := &Round{
		Scores:           []int{4, 2, 0, 1},
		Dealer:           1,
		Wind:             DirectionNorth,
		ReservedDuration: 2 * time.Second,
	}
	r.Start(0)
	_ = r.Discard(1, time.Time{}, TileBamboo1)
	t.Run("view from seat", func(t *testing.T) {
		seat := 1
		view := r.View(seat)
		assert.Equal(
			t,
			RoundView{
				Seat:   seat,
				Scores: r.Scores,
				Hand:   r.Hands[seat],
				Hands: []HandView{
					{Flowers: []Tile{}, Revealed: []Meld{}, Concealed: 13},
					{Flowers: []Tile{"06兰", "12冬"}, Revealed: []Meld{}, Concealed: 13},
					{Flowers: []Tile{"07菊"}, Revealed: []Meld{}, Concealed: 13},
				},
				DrawsLeft: len(r.Wall) - 16,
				Discards:  r.Discards,
				Wind:      r.Wind,
				Dealer:    r.Dealer,
				Turn:      r.Turn,
				Phase:     r.Phase,
				Events: []EventView{{
					Type:  EventDiscard,
					Seat:  1,
					Tiles: []Tile{TileBamboo1},
				}},

				Result:           r.Result,
				LastDiscardTime:  r.LastDiscardTime,
				ReservedDuration: r.ReservedDuration,
			},
			view,
		)
	})
	t.Run("bystander's view", func(t *testing.T) {
		view := r.View(-1)
		t.Logf("%#v", view.Hands)
		assert.Equal(
			t,
			RoundView{
				Scores: r.Scores,
				Hands: []HandView{
					{Flowers: []Tile{"07菊"}, Revealed: []Meld{}, Concealed: 13},
					{Flowers: []Tile{}, Revealed: []Meld{}, Concealed: 13},
					{Flowers: []Tile{}, Revealed: []Meld{}, Concealed: 13},
					{Flowers: []Tile{"06兰", "12冬"}, Revealed: []Meld{}, Concealed: 13},
				},
				DrawsLeft: len(r.Wall) - 16,
				Discards:  r.Discards,
				Wind:      r.Wind,
				Dealer:    r.Dealer,
				Turn:      r.Turn,
				Phase:     r.Phase,
				Events: []EventView{{
					Type:  EventDiscard,
					Seat:  1,
					Tiles: []Tile{TileBamboo1},
				}},
				Result:           r.Result,
				LastDiscardTime:  r.LastDiscardTime,
				ReservedDuration: r.ReservedDuration,
			},
			view,
		)
	})
}

func Test_newWall(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	got := newWall(r)
	want := []Tile{"38八万", "35五万", "27六索", "44红中", "22一索", "34四万", "35五万", "20八筒", "37七万", "13一筒", "43北风", "26五索", "21九筒", "25四索", "42西风", "17五筒", "38八万", "36六万", "16四筒", "43北风", "20八筒", "22一索", "37七万", "25四索", "42西风", "30九索", "19七筒", "06兰", "27六索", "07菊", "40东风", "32二万", "29八索", "36六万", "34四万", "46白板", "32二万", "15三筒", "17五筒", "37七万", "42西风", "14二筒", "43北风", "20八筒", "28七索", "45青发", "17五筒", "36六万", "34四万", "14二筒", "12冬", "46白板", "22一索", "40东风", "37七万", "28七索", "29八索", "16四筒", "39九万", "13一筒", "24三索", "01猫", "27六索", "40东风", "41南风", "34四万", "24三索", "31一万", "31一万", "25四索", "13一筒", "26五索", "15三筒", "14二筒", "18六筒", "24三索", "11秋", "19七筒", "45青发", "41南风", "44红中", "39九万", "27六索", "26五索", "10夏", "15三筒", "21九筒", "36六万", "41南风", "33三万", "29八索", "23二索", "28七索", "04蜈蚣", "32二万", "38八万", "29八索", "05梅", "39九万", "21九筒", "46白板", "33三万", "09春", "32二万", "25四索", "30九索", "39九万", "23二索", "02老鼠", "24三索", "44红中", "28七索", "45青发", "18六筒", "31一万", "14二筒", "43北风", "13一筒", "45青发", "30九索", "18六筒", "22一索", "31一万", "16四筒", "17五筒", "26五索", "23二索", "21九筒", "35五万", "42西风", "03公鸡", "35五万", "18六筒", "30九索", "46白板", "38八万", "40东风", "19七筒", "15三筒", "41南风", "33三万", "16四筒", "20八筒", "23二索", "08竹", "33三万", "19七筒", "44红中"}
	assert.Equal(t, want, got)
}

func TestRound_distributeTiles(t *testing.T) {
	r := &Round{
		Dealer: 1,
		Wall: []Tile{
			"38八万", "35五万", "27六索", "44红中", // dealer (1) draws first
			"22一索", "34四万", "35五万", "20八筒",
			"37七万", "13一筒", "43北风", "26五索",
			"21九筒", "25四索", "42西风", "17五筒",

			"38八万", "36六万", "16四筒", "43北风",
			"20八筒", "22一索", "37七万", "25四索",
			"42西风", "30九索", "19七筒", "06兰", // 3 draws a flower
			"27六索", "07菊", "40东风", "32二万", // 0 draws a flower

			"29八索", "36六万", "34四万", "46白板",
			"32二万", "15三筒", "17五筒", "37七万",
			"42西风", "14二筒", "43北风", "20八筒",
			"28七索", "45青发", "17五筒", "36六万",

			"34四万",
			"14二筒",
			"12冬", // 3 draws another flower
			"46白板",

			"22一索",

			"40东风", "37七万", "28七索", "29八索", "16四筒", "39九万", "13一筒", "24三索", "44红中", "27六索", "40东风", "41南风", "34四万", "24三索", "31一万", "31一万", "25四索", "13一筒", "26五索", "15三筒", "14二筒", "18六筒", "24三索", "11秋", "19七筒", "45青发", "41南风", "44红中", "39九万", "27六索", "26五索", "10夏", "15三筒", "21九筒", "36六万", "41南风", "33三万", "29八索", "23二索", "28七索", "04蜈蚣", "32二万", "38八万", "29八索", "05梅", "39九万", "21九筒", "46白板", "33三万", "09春", "32二万", "25四索", "30九索", "39九万", "23二索", "02老鼠", "24三索", "44红中", "28七索", "45青发", "18六筒", "31一万", "14二筒", "43北风", "13一筒", "45青发", "30九索", "18六筒", "22一索", "31一万", "16四筒", "17五筒", "26五索", "23二索", "21九筒", "35五万", "42西风", "03公鸡", "35五万", "18六筒", "30九索", "46白板", "38八万", "40东风", "19七筒", "15三筒", "41南风", "33三万", "16四筒", "20八筒",

			"23二索",        // 3 replaces the fourth flower
			"08竹",         // 3 replaces the third flower and gets a fourth flower
			"33三万",        // 0 replaces one tile
			"19七筒", "01猫", // 3 replaces two tiles and gets a third flower
		},
	}
	r.distributeTiles()
	assert.Equal(t,
		NewTileBag([]Tile{"38八万", "35五万", "27六索", "44红中", "38八万", "36六万", "16四筒", "43北风", "29八索", "36六万", "34四万", "46白板", "34四万", "22一索"}),
		r.Hands[1].Concealed)
	assert.Equal(t,
		NewTileBag([]Tile{"22一索", "34四万", "35五万", "20八筒", "20八筒", "22一索", "37七万", "25四索", "32二万", "15三筒", "17五筒", "37七万", "14二筒"}),
		r.Hands[2].Concealed)
	assert.Equal(t,
		NewTileBag([]Tile{
			"37七万", "13一筒", "43北风", "26五索",
			"42西风", "30九索", "19七筒",
			"42西风", "14二筒", "43北风", "20八筒",
			"19七筒", "23二索", // replaced tiles
		}),
		r.Hands[3].Concealed)
	assert.ElementsMatch(t, []Tile{"06兰", "12冬", "01猫", "08竹"}, r.Hands[3].Flowers)
	assert.Equal(t,
		NewTileBag([]Tile{
			"21九筒", "25四索", "42西风", "17五筒",
			"27六索", "40东风", "32二万",
			"28七索", "45青发", "17五筒", "36六万",
			"46白板",
			"33三万", // replaced tile
		}),
		r.Hands[0].Concealed)
	assert.ElementsMatch(t, []Tile{"07菊"}, r.Hands[0].Flowers)
	assert.Equal(t,
		[]Tile{"40东风", "37七万", "28七索", "29八索", "16四筒", "39九万", "13一筒", "24三索", "44红中", "27六索", "40东风", "41南风", "34四万", "24三索", "31一万", "31一万", "25四索", "13一筒", "26五索", "15三筒", "14二筒", "18六筒", "24三索", "11秋", "19七筒", "45青发", "41南风", "44红中", "39九万", "27六索", "26五索", "10夏", "15三筒", "21九筒", "36六万", "41南风", "33三万", "29八索", "23二索", "28七索", "04蜈蚣", "32二万", "38八万", "29八索", "05梅", "39九万", "21九筒", "46白板", "33三万", "09春", "32二万", "25四索", "30九索", "39九万", "23二索", "02老鼠", "24三索", "44红中", "28七索", "45青发", "18六筒", "31一万", "14二筒", "43北风", "13一筒", "45青发", "30九索", "18六筒", "22一索", "31一万", "16四筒", "17五筒", "26五索", "23二索", "21九筒", "35五万", "42西风", "03公鸡", "35五万", "18六筒", "30九索", "46白板", "38八万", "40东风", "19七筒", "15三筒", "41南风", "33三万", "16四筒", "20八筒"},
		r.Wall)
}

func TestRound_Start(t *testing.T) {
	r := new(Round)
	r.Start(0)
	assert.Equal(t, r.Dealer, r.Turn)
	assert.Equal(t, r.Phase, PhaseDiscard)
}

func TestRound_End(t *testing.T) {
	t.Run("can only end round during own turn", func(t *testing.T) {
		r := &Round{
			Turn: 1,
		}
		err := r.End(0, time.Now())
		assert.EqualError(t, err, "wrong turn")
	})
	t.Run("can only end round during discard phase", func(t *testing.T) {
		r := &Round{
			Turn:  0,
			Phase: PhaseDraw,
		}
		err := r.End(0, time.Now())
		assert.EqualError(t, err, "wrong phase")
	})
	t.Run("can only end round when there are less than 16 tiles in the wall", func(t *testing.T) {
		r := &Round{
			Turn:  0,
			Phase: PhaseDiscard,
			Wall: []Tile{
				"38八万", "35五万", "27六索", "44红中",
				"22一索", "34四万", "35五万", "20八筒",
				"37七万", "13一筒", "43北风", "26五索",
				"21九筒", "25四索", "42西风", "17五筒",
			},
		}
		err := r.End(0, time.Now())
		assert.EqualError(t, err, "some draws remaining")
	})
	t.Run("successfully ending round", func(t *testing.T) {
		r := &Round{
			Turn:  0,
			Phase: PhaseDiscard,
		}
		err := r.End(0, time.Now())
		assert.NoError(t, err)
		assert.Equal(t, PhaseFinished, r.Phase)
		assert.Equal(t, Result{
			Dealer: r.Dealer,
			Wind:   r.Wind,
			Winner: -1,
		}, r.Result)
	})
}
