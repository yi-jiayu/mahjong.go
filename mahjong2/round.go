package mahjong2

import (
	"errors"
	"sort"
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

func (r *Round) Draw(seat int, t time.Time) (drawn Tile, flowers []Tile, err error) {
	if r.Turn != seat {
		err = errors.New("wrong turn")
		return
	}
	if r.Phase != PhaseDraw {
		err = errors.New("wrong phase")
		return
	}
	if t.Before(r.LastDiscardTime.Add(r.ReservedDuration)) {
		err = errors.New("cannot draw during reserved duration")
		return
	}
	drawn = r.drawFront()
	flowers = make([]Tile, 0)
	for isFlower(drawn) {
		flowers = append(flowers, drawn)
		drawn = r.drawBack()
	}
	hand := &r.Hands[seat]
	hand.Concealed.Add(drawn)
	hand.Flowers = append(hand.Flowers, flowers...)
	r.Phase = PhaseDiscard
	return
}

func (r *Round) Discard(seat int, t time.Time, tile Tile) error {
	if seat != r.Turn {
		return errors.New("wrong turn")
	}
	if r.Phase != PhaseDiscard {
		return errors.New("wrong phase")
	}
	if !r.Hands[seat].Concealed.Contains(tile) {
		return errors.New("missing tiles")
	}
	r.Hands[seat].Concealed.Remove(tile)
	r.Discards = append(r.Discards, tile)
	r.Turn = (r.Turn + 1) % 4
	r.Phase = PhaseDraw
	return nil
}

func (r *Round) Chi(seat int, t time.Time, tile1, tile2 Tile) error {
	if r.Turn != seat {
		return errors.New("wrong turn")
	}
	if r.Phase != PhaseDraw {
		return errors.New("wrong phase")
	}
	if len(r.Discards) == 0 {
		return errors.New("no discards")
	}
	tile0 := r.lastDiscard()
	if !isValidSequence(tile0, tile1, tile2) {
		return errors.New("invalid sequence")

	}
	hand := &r.Hands[seat]
	if !hand.Concealed.Contains(tile1) || !hand.Concealed.Contains(tile2) {
		return errors.New("missing tiles")
	}
	if t.Before(r.LastDiscardTime.Add(r.ReservedDuration)) {
		return errors.New("cannot chi during reserved duration")
	}
	hand.Concealed.Remove(tile1)
	hand.Concealed.Remove(tile2)
	r.popLastDiscard()
	seq := []Tile{tile0, tile1, tile2}
	sort.Slice(seq, func(i, j int) bool {
		return seq[i] < seq[j]
	})
	hand.Revealed = append(hand.Revealed, Meld{
		Type:  MeldChi,
		Tiles: seq,
	})
	r.Phase = PhaseDiscard
	return nil
}

func (r *Round) Pong(seat int, t time.Time) error {
	if seat == r.previousTurn() {
		return errors.New("wrong turn")
	}
	if r.Phase != PhaseDraw {
		return errors.New("wrong phase")
	}
	if len(r.Discards) == 0 {
		return errors.New("no discards")
	}
	hand := &r.Hands[seat]
	if hand.Concealed.Count(r.lastDiscard()) < 2 {
		return errors.New("missing tiles")
	}
	tile := r.popLastDiscard()
	hand.Concealed.RemoveN(tile, 2)
	hand.Revealed = append(hand.Revealed, Meld{
		Type:  MeldPong,
		Tiles: []Tile{tile},
	})
	r.Turn = seat
	r.Phase = PhaseDiscard
	return nil
}

func (r *Round) GangFromDiscard(seat int, t time.Time) (replacement Tile, flowers []Tile, err error) {
	if seat == r.previousTurn() {
		err = errors.New("wrong turn")
		return
	}
	if r.Phase != PhaseDraw {
		err = errors.New("wrong phase")
		return
	}
	if len(r.Discards) == 0 {
		err = errors.New("no discards")
		return
	}
	hand := &r.Hands[seat]
	if hand.Concealed.Count(r.lastDiscard()) < 3 {
		err = errors.New("missing tiles")
		return
	}
	tile := r.popLastDiscard()
	hand.Concealed.RemoveN(tile, 3)
	hand.Revealed = append(hand.Revealed, Meld{
		Type:  MeldGang,
		Tiles: []Tile{tile},
	})
	replacement, flowers = r.replaceTile()
	hand.Flowers = append(hand.Flowers, flowers...)
	hand.Concealed.Add(replacement)
	r.Turn = seat
	r.Phase = PhaseDiscard
	return
}

func (r *Round) GangFromHand(seat int, t time.Time, tile Tile) (replacement Tile, flowers []Tile, err error) {
	if seat != r.Turn {
		err = errors.New("wrong turn")
		return
	}
	if r.Phase != PhaseDiscard {
		err = errors.New("wrong phase")
		return
	}
	hand := &r.Hands[seat]
	if hand.Concealed.Count(tile) == 4 {
		hand.Concealed.RemoveN(tile, 4)
		hand.Revealed = append(hand.Revealed, Meld{
			Type:  MeldGang,
			Tiles: []Tile{tile},
		})
		replacement, flowers = r.replaceTile()
		hand.Flowers = append(hand.Flowers, flowers...)
		hand.Concealed.Add(replacement)
		return
	}
	for i, meld := range hand.Revealed {
		if meld.Type == MeldPong && meld.Tiles[0] == tile && hand.Concealed.Count(tile) > 0 {
			hand.Concealed.Remove(tile)
			hand.Revealed[i].Type = MeldGang
			replacement, flowers = r.replaceTile()
			hand.Flowers = append(hand.Flowers, flowers...)
			hand.Concealed.Add(replacement)
			return
		}
	}
	err = errors.New("missing tiles")
	return
}

func (r *Round) Hu(seat int, t time.Time) error {
	if seat == r.previousTurn() {
		return errors.New("wrong turn")
	}
	if r.Turn != seat && r.Phase == PhaseDiscard {
		return errors.New("wrong turn")
	}
	if r.Turn == seat && r.Phase == PhaseDraw {
		return errors.New("wrong phase")
	}
	// temporarily add the last discard to the player's hand
	// if trying to hu from a discard
	if r.Phase == PhaseDraw {
		r.Hands[seat].Concealed.Add(r.lastDiscard())
	}
	winningHands := search(r.Hands[seat].Concealed)
	if len(winningHands) == 0 {
		if r.Phase == PhaseDraw {
			r.Hands[seat].Concealed.Remove(r.lastDiscard())
		}
		return errors.New("missing tiles")
	}
	// actually remove the last discard if the player has a winning hand
	if r.Phase == PhaseDraw {
		r.popLastDiscard()
	}
	// for now, take the first winning hand
	// ideally we will take the highest scoring hand
	winningHand := Melds(append(winningHands[0], r.Hands[seat].Revealed...))
	sort.Sort(winningHand)
	r.Result = Result{
		Dealer:       r.Dealer,
		Wind:         r.Wind,
		Winner:       seat,
		WinningTiles: append(r.Hands[seat].Flowers, winningHand.Tiles()...),
	}
	r.Phase = PhaseFinished
	return nil
}
