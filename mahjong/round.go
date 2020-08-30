package mahjong

import (
	"errors"
	"math/rand"
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

	LastActionTime   time.Time
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

func (r *Round) drawFrontN(n int) []Tile {
	draws := r.Wall[:n]
	r.Wall = r.Wall[n:]
	return draws
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

func (r *Round) seatWind(seat int) Direction {
	return Direction((seat - r.Dealer + 4) % 4)
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
	if t.Before(r.LastActionTime.Add(r.ReservedDuration)) {
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
	r.Events = append(r.Events, DrawEvent{
		Seat:    seat,
		Time:    t,
		Tile:    drawn,
		Flowers: flowers,
	})
	r.LastActionTime = t
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
	r.Events = append(r.Events, DiscardEvent{
		Seat: seat,
		Time: t,
		Tile: tile,
	})
	r.LastActionTime = t
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
	if t.Before(r.LastActionTime.Add(r.ReservedDuration)) {
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
	r.Events = append(r.Events, ChiEvent{
		Seat:        seat,
		Time:        t,
		LastDiscard: tile0,
		Tiles:       [2]Tile{tile1, tile2},
	})
	r.LastActionTime = t
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
	r.Events = append(r.Events, PongEvent{
		Seat:         seat,
		Time:         t,
		Tile:         tile,
		PreviousTurn: r.Turn,
	})
	r.Turn = seat
	r.Phase = PhaseDiscard
	r.LastActionTime = t
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
	r.Events = append(r.Events, GangEvent{
		Seat: seat,
		Time: t,
		Tile: tile,
	})
	r.Turn = seat
	r.Phase = PhaseDiscard
	r.LastActionTime = t
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
	if hand.Concealed.Count(tile) > 3 {
		hand.Concealed.RemoveN(tile, 4)
		hand.Revealed = append(hand.Revealed, Meld{
			Type:  MeldGang,
			Tiles: []Tile{tile},
		})
		replacement, flowers = r.replaceTile()
		hand.Flowers = append(hand.Flowers, flowers...)
		hand.Concealed.Add(replacement)
		r.Events = append(r.Events, GangEvent{
			Seat: seat,
			Time: t,
			Tile: tile,
		})
		r.LastActionTime = t
		return
	}
	for i, meld := range hand.Revealed {
		if meld.Type == MeldPong && meld.Tiles[0] == tile && hand.Concealed.Count(tile) > 0 {
			hand.Concealed.Remove(tile)
			hand.Revealed[i].Type = MeldGang
			replacement, flowers = r.replaceTile()
			hand.Flowers = append(hand.Flowers, flowers...)
			hand.Concealed.Add(replacement)
			r.Events = append(r.Events, GangEvent{
				Seat: seat,
				Time: t,
				Tile: tile,
			})
			r.LastActionTime = t
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
	r.Hands[seat].Revealed = append(r.Hands[seat].Revealed, winningHands[0]...)
	r.Hands[seat].Concealed = TileBag{}
	sort.Sort(Melds(r.Hands[seat].Revealed))
	r.Result = Result{
		Dealer:       r.Dealer,
		Wind:         r.Wind,
		Winner:       seat,
		WinningTiles: append(r.Hands[seat].Flowers, Melds(r.Hands[seat].Revealed).Tiles()...),
	}
	r.LastActionTime = t
	r.Events = append(r.Events, HuEvent{
		Seat: seat,
		Time: t,
	})
	r.Phase = PhaseFinished
	return nil
}

func (r *Round) distributeTiles() {
	r.Hands = []Hand{
		{
			Flowers:   []Tile{},
			Revealed:  []Meld{},
			Concealed: TileBag{},
		},
		{
			Flowers:   []Tile{},
			Revealed:  []Meld{},
			Concealed: TileBag{},
		},
		{
			Flowers:   []Tile{},
			Revealed:  []Meld{},
			Concealed: TileBag{},
		},
		{
			Flowers:   []Tile{},
			Revealed:  []Meld{},
			Concealed: TileBag{},
		},
	}
	order := []int{r.Dealer, (r.Dealer + 1) % 4, (r.Dealer + 2) % 4, (r.Dealer + 3) % 4}
	// draw 4 tiles 3 times
	for i := 0; i < 3; i++ {
		var draws []Tile
		for _, seat := range order {
			draws = r.drawFrontN(4)
			r.Hands[seat].Concealed.Add(draws...)
		}
	}
	// draw one tile
	var draw Tile
	for _, seat := range order {
		draw = r.drawFront()
		r.Hands[seat].Concealed.Add(draw)
	}
	// dealer draws one extra tile
	draw = r.drawFront()
	r.Hands[r.Dealer].Concealed.Add(draw)
	// replace flowers
	for len(order) > 0 {
		seat := order[0]
		order = order[1:]
		mustReplaceAgain := false
		var flowers, replacements []Tile
		for tile := range r.Hands[seat].Concealed {
			if isFlower(tile) {
				draw = r.drawBack()
				if isFlower(draw) {
					mustReplaceAgain = true
				}
				flowers = append(flowers, tile)
				replacements = append(replacements, draw)
			}
		}
		if mustReplaceAgain {
			order = append(order, seat)
		}
		sort.Slice(flowers, func(i, j int) bool {
			return flowers[i] < flowers[j]
		})
		r.Hands[seat].Flowers = append(r.Hands[seat].Flowers, flowers...)
		r.Hands[seat].Concealed.Remove(flowers...)
		r.Hands[seat].Concealed.Add(replacements...)
	}
	return
}

func (r *Round) Start(seed int64, t time.Time) {
	r.Wall = newWall(rand.New(rand.NewSource(seed)))
	r.distributeTiles()
	r.Turn = r.Dealer
	r.Phase = PhaseDiscard
	r.Discards = []Tile{}
	r.LastActionTime = t
	r.Events = []Event{StartEvent{Time: t}}
}

// Next returns a new round, setting the dealer and the prevailing wind
// depending on the outcome of this round.
func (r *Round) Next() (*Round, error) {
	if r.Phase != PhaseFinished {
		return nil, errors.New("unfinished")
	}
	dealer := r.Dealer
	wind := r.Wind
	if r.Result.Winner != dealer {
		if dealer == 3 && wind == DirectionNorth {
			return nil, errors.New("no more rounds")
		}
		dealer = (r.Dealer + 1) % 4
		if dealer == 0 {
			wind++
		}
	}
	return &Round{
		Dealer: dealer,
		Wind:   wind,
	}, nil
}

// End ends a round in a draw. Only the player who drew the last available tile
// from the wall may initiate this action.
func (r *Round) End(seat int, t time.Time) error {
	if r.Turn != seat {
		return errors.New("wrong turn")
	}
	if r.Phase != PhaseDiscard {
		return errors.New("wrong phase")
	}
	if len(r.Wall) >= 16 {
		return errors.New("some draws remaining")
	}
	r.Phase = PhaseFinished
	r.Result = Result{
		Dealer: r.Dealer,
		Wind:   r.Wind,
		Winner: -1,
	}
	r.LastActionTime = t
	r.Events = append(r.Events, EndEvent{
		Seat: seat,
		Time: t,
	})
	return nil
}

// View returns a view of a round from a certain seat. Values of seat outside
// of [0, 3] will return a bystander's view of the round.
func (r *Round) View(seat int) RoundView {
	events := make([]EventView, len(r.Events))
	for i, event := range r.Events {
		events[i] = event.View()
	}
	hands := make([]Hand, 4)
	for i, hand := range r.Hands {
		if seat == i {
			hands[i] = hand
		} else {
			hands[i] = hand.View()
		}
	}
	return RoundView{
		Seat:             seat,
		Scores:           r.Scores,
		Hands:            hands,
		DrawsLeft:        len(r.Wall) - 16,
		Discards:         r.Discards,
		Wind:             r.Wind,
		Dealer:           r.Dealer,
		Turn:             r.Turn,
		Phase:            r.Phase,
		Events:           events,
		Result:           r.Result,
		LastActionTime:   r.LastActionTime.UnixNano() / 1e6,
		ReservedDuration: r.ReservedDuration.Milliseconds(),
	}
}

func newWall(r *rand.Rand) []Tile {
	var wall []Tile
	wall = append(wall, flowerTiles...)
	for _, tile := range suitedTiles {
		wall = append(wall, tile, tile, tile, tile)
	}
	r.Shuffle(len(wall), func(i, j int) {
		wall[i], wall[j] = wall[j], wall[i]
	})
	return wall
}
