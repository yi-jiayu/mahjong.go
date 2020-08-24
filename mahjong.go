package mahjong

import (
	"errors"
	"math/rand"
	"sort"
	"strconv"
	"time"
)

type Tile string

const (
	TileCat          Tile = "01猫"
	TileRat          Tile = "02老鼠"
	TileRooster      Tile = "03公鸡"
	TileCentipede    Tile = "04蜈蚣"
	TileGentlemen1   Tile = "05梅"
	TileGentlemen2   Tile = "06兰"
	TileGentlemen3   Tile = "07菊"
	TileGentlemen4   Tile = "08竹"
	TileSeasons1     Tile = "09春"
	TileSeasons2     Tile = "10夏"
	TileSeasons3     Tile = "11秋"
	TileSeasons4     Tile = "12冬"
	TileDots1        Tile = "13一筒"
	TileDots2        Tile = "14二筒"
	TileDots3        Tile = "15三筒"
	TileDots4        Tile = "16四筒"
	TileDots5        Tile = "17五筒"
	TileDots6        Tile = "18六筒"
	TileDots7        Tile = "19七筒"
	TileDots8        Tile = "20八筒"
	TileDots9        Tile = "21九筒"
	TileBamboo1      Tile = "22一索"
	TileBamboo2      Tile = "23二索"
	TileBamboo3      Tile = "24三索"
	TileBamboo4      Tile = "25四索"
	TileBamboo5      Tile = "26五索"
	TileBamboo6      Tile = "27六索"
	TileBamboo7      Tile = "28七索"
	TileBamboo8      Tile = "29八索"
	TileBamboo9      Tile = "30九索"
	TileCharacters1  Tile = "31一万"
	TileCharacters2  Tile = "32二万"
	TileCharacters3  Tile = "33三万"
	TileCharacters4  Tile = "34四万"
	TileCharacters5  Tile = "35五万"
	TileCharacters6  Tile = "36六万"
	TileCharacters7  Tile = "37七万"
	TileCharacters8  Tile = "38八万"
	TileCharacters9  Tile = "39九万"
	TileWindsEast    Tile = "40东风"
	TileWindsSouth   Tile = "41南风"
	TileWindsWest    Tile = "42西风"
	TileWindsNorth   Tile = "43北风"
	TileDragonsRed   Tile = "44红中"
	TileDragonsGreen Tile = "45青发"
	TileDragonsWhite Tile = "46白板"
)

type Direction int

const (
	DirectionEast Direction = iota
	DirectionSouth
	DirectionWest
	DirectionNorth
)

type Action string

const (
	ActionDraw     Action = "draw"
	ActionDiscard  Action = "discard"
	ActionGameOver Action = "game over"
)

var (
	FlowerTiles = []Tile{"01猫", "02老鼠", "03公鸡", "04蜈蚣", "05梅", "06兰", "07菊", "08竹", "09春", "10夏", "11秋", "12冬"}
	NormalTiles = []Tile{"13一筒", "14二筒", "15三筒", "16四筒", "17五筒", "18六筒", "19七筒", "20八筒", "21九筒", "22一索", "23二索", "24三索", "25四索", "26五索", "27六索", "28七索", "29八索", "30九索", "31一万", "32二万", "33三万", "34四万", "35五万", "36六万", "37七万", "38八万", "39九万", "40东风", "41南风", "42西风", "43北风", "44红中", "45青发", "46白板"}
)

var validSequences = [][3]Tile{
	{TileBamboo1, TileBamboo2, TileBamboo3},
	{TileBamboo2, TileBamboo3, TileBamboo4},
	{TileBamboo3, TileBamboo4, TileBamboo5},
	{TileBamboo4, TileBamboo5, TileBamboo6},
	{TileBamboo5, TileBamboo6, TileBamboo7},
	{TileBamboo6, TileBamboo7, TileBamboo8},
	{TileBamboo7, TileBamboo8, TileBamboo9},
	{TileDots1, TileDots2, TileDots3},
	{TileDots2, TileDots3, TileDots4},
	{TileDots3, TileDots4, TileDots5},
	{TileDots4, TileDots5, TileDots6},
	{TileDots5, TileDots6, TileDots7},
	{TileDots6, TileDots7, TileDots8},
	{TileDots7, TileDots8, TileDots9},
	{TileCharacters1, TileCharacters2, TileCharacters3},
	{TileCharacters2, TileCharacters3, TileCharacters4},
	{TileCharacters3, TileCharacters4, TileCharacters5},
	{TileCharacters4, TileCharacters5, TileCharacters6},
	{TileCharacters5, TileCharacters6, TileCharacters7},
	{TileCharacters6, TileCharacters7, TileCharacters8},
	{TileCharacters7, TileCharacters8, TileCharacters9},
}

// MinTilesLeft is the minimum number of tiles left before the game is considered a draw.
const MinTilesLeft = 16

type Hand struct {
	Flowers   []Tile   `json:"flowers"`
	Revealed  [][]Tile `json:"revealed"`
	Concealed []Tile   `json:"concealed"`
}

// Masked returns a copy of a Hand with concealed tiles replaced with empty strings.
func (h Hand) Masked() Hand {
	return Hand{
		Flowers:   h.Flowers,
		Revealed:  h.Revealed,
		Concealed: make([]Tile, len(h.Concealed)),
	}
}

type EventType string

const (
	EventDraw    EventType = "draw"
	EventDiscard EventType = "discard"
	EventChi     EventType = "chi"
	EventPong    EventType = "pong"
	EventGang    EventType = "gang"
)

type Event struct {
	Type  EventType `json:"type"`
	Seat  Direction `json:"seat"`
	Time  time.Time `json:"time"`
	Tiles []Tile    `json:"tiles"`
}

type Round struct {
	Wall          []Tile
	Discards      []Tile
	Hands         [4]Hand
	CurrentTurn   Direction
	CurrentAction Action

	// PongDuration is an amount of time after the last discard that is reserved for ponging,
	// which the next player has to wait before he can draw or chi.
	PongDuration time.Duration

	// LastDiscardTime is the time of the last discard.
	LastDiscardTime time.Time

	// Events describes everything that happened this round.
	Events []Event
}

type MeldType int

const (
	MeldChow MeldType = iota
	MeldPeng
	MeldKong
	MeldEyes
)

type Meld struct {
	Type  MeldType
	Tiles []Tile
}

func newWall(r *rand.Rand) []Tile {
	var wall []Tile
	wall = append(wall, FlowerTiles...)
	for _, tile := range NormalTiles {
		wall = append(wall, tile, tile, tile, tile)
	}
	r.Shuffle(len(wall), func(i, j int) {
		wall[i], wall[j] = wall[j], wall[i]
	})
	return wall
}

func drawFront(wall []Tile) (Tile, []Tile) {
	drawn := wall[0]
	wall = wall[1:]
	return drawn, wall
}

func drawFrontN(wall []Tile, n int) ([]Tile, []Tile) {
	drawn := wall[:n]
	wall = wall[n:]
	return drawn, wall
}

func drawBack(wall []Tile) (Tile, []Tile) {
	drawn := wall[len(wall)-1]
	wall = wall[:len(wall)-1]
	return drawn, wall
}

func contains(tiles []Tile, tile Tile) bool {
	for _, t := range tiles {
		if t == tile {
			return true
		}
	}
	return false
}

func isFlower(tile Tile) bool {
	return contains(FlowerTiles, tile)
}

func distributeTiles(wall []Tile) ([4]Hand, []Tile) {
	hands := [4]Hand{
		{
			Flowers:   []Tile{},
			Revealed:  [][]Tile{},
			Concealed: []Tile{},
		},
		{
			Flowers:   []Tile{},
			Revealed:  [][]Tile{},
			Concealed: []Tile{},
		},
		{
			Flowers:   []Tile{},
			Revealed:  [][]Tile{},
			Concealed: []Tile{},
		},
		{
			Flowers:   []Tile{},
			Revealed:  [][]Tile{},
			Concealed: []Tile{},
		},
	}
	order := []Direction{DirectionEast, DirectionSouth, DirectionWest, DirectionNorth}
	// draw 4 tiles 3 times
	for i := 0; i < 3; i++ {
		var draws []Tile
		for _, seat := range order {
			draws, wall = drawFrontN(wall, 4)
			hands[seat].Concealed = append(hands[seat].Concealed, draws...)
		}
	}
	// draw one tile
	var draw Tile
	for _, seat := range order {
		draw, wall = drawFront(wall)
		hands[seat].Concealed = append(hands[seat].Concealed, draw)
	}
	// dealer draws one extra tile
	draw, wall = drawFront(wall)
	hands[DirectionEast].Concealed = append(hands[DirectionEast].Concealed, draw)
	// replace flowers
	replacementOrder := order
	for len(replacementOrder) > 0 {
		seat := replacementOrder[0]
		replacementOrder = replacementOrder[1:]
		i := 0
		mustReplaceAgain := false
		for _, tile := range hands[seat].Concealed {
			if isFlower(tile) {
				hands[seat].Flowers = append(hands[seat].Flowers, tile)
				draw, wall = drawBack(wall)
				if isFlower(draw) {
					mustReplaceAgain = true
				}
				hands[seat].Concealed[i] = draw
			} else {
				hands[seat].Concealed[i] = tile
			}
			i++
			if mustReplaceAgain {
				replacementOrder = append(replacementOrder, seat)
			}
		}
		hands[seat].Concealed = hands[seat].Concealed[:i]
	}
	return hands, wall
}

func removeTiles(tiles []Tile, tile Tile, count int) ([]Tile, int) {
	i := 0
	removedCount := 0
	for _, t := range tiles {
		if t == tile && removedCount < count {
			removedCount++
		} else {
			tiles[i] = t
			i++
		}
	}
	tiles = tiles[:i]
	return tiles, removedCount
}

func removeTile(tiles []Tile, tile Tile) ([]Tile, bool) {
	tiles, removedCount := removeTiles(tiles, tile, 1)
	return tiles, removedCount > 0
}

func (r *Round) Discard(seat Direction, tile Tile, t time.Time) error {
	if r.CurrentTurn != seat {
		return errors.New("not your turn")
	}
	if r.CurrentAction != ActionDiscard {
		return errors.New("not time to discard")
	}
	if len(r.Wall) < MinTilesLeft {
		return errors.New("cannot discard on last round")
	}
	remaining, ok := removeTile(r.Hands[seat].Concealed, tile)
	if !ok {
		return errors.New("no such tile")
	}
	r.Hands[seat].Concealed = remaining
	r.Discards = append(r.Discards, tile)
	r.CurrentTurn = (seat + 1) % 4
	r.CurrentAction = ActionDraw
	r.LastDiscardTime = t
	r.Events = append(r.Events, Event{
		Type:  EventDiscard,
		Seat:  seat,
		Time:  t,
		Tiles: []Tile{tile},
	})
	return nil
}

func validSequence(seq [3]Tile) bool {
	sort.Slice(seq[:], func(i, j int) bool {
		return seq[i] < seq[j]
	})
	for _, valid := range validSequences {
		if seq == valid {
			return true
		}
	}
	return false
}

func (r *Round) Chow(seat Direction, tile1, tile2 Tile, t time.Time) error {
	if r.CurrentTurn != seat {
		return errors.New("not your turn")
	}
	if r.CurrentAction != ActionDraw {
		return errors.New("not time to chow")
	}
	if t.Before(r.LastDiscardTime.Add(r.PongDuration)) {
		return errors.New("waiting for other players to pong")
	}
	if !contains(r.Hands[seat].Concealed, tile1) || !contains(r.Hands[seat].Concealed, tile2) {
		return errors.New("no such tile")
	}
	seq := [3]Tile{tile1, tile2, r.lastDiscard()}
	if !validSequence(seq) {
		return errors.New("invalid sequence")
	}
	r.Hands[seat].Concealed, _ = removeTile(r.Hands[seat].Concealed, tile1)
	r.Hands[seat].Concealed, _ = removeTile(r.Hands[seat].Concealed, tile2)
	r.Hands[seat].Revealed = append(r.Hands[seat].Revealed, seq[:])
	r.Discards = r.Discards[:len(r.Discards)-1]
	r.CurrentAction = ActionDiscard
	r.Events = append(r.Events, Event{
		Type:  EventChi,
		Seat:  seat,
		Time:  t,
		Tiles: seq[:],
	})
	return nil
}

func (r *Round) PreviousTurn() Direction {
	return (r.CurrentTurn + 3) % 4
}

func countTiles(tiles []Tile, tile Tile) int {
	count := 0
	for _, t := range tiles {
		if t == tile {
			count++
		}
	}
	return count
}

func (r *Round) Peng(seat Direction, tile Tile) error {
	if seat == r.PreviousTurn() {
		return errors.New("wrong turn")
	}
	if r.CurrentAction != ActionDraw {
		return errors.New("wrong action")
	}
	if countTiles(r.Hands[seat].Concealed, tile) < 2 {
		return errors.New("not enough tiles")
	}
	r.Hands[seat].Concealed, _ = removeTiles(r.Hands[seat].Concealed, tile, 2)
	r.Hands[seat].Revealed = append(r.Hands[seat].Revealed, []Tile{tile, tile, tile})
	r.Discards = r.Discards[:len(r.Discards)-1]
	r.CurrentAction = ActionDiscard
	r.CurrentTurn = seat
	r.Events = append(r.Events, Event{
		Type:  EventPong,
		Seat:  seat,
		Tiles: []Tile{tile},
	})
	return nil
}

// Draw draws a new tile for seat and returns the drawn tile and any flowers drawn in the process.
func (r *Round) Draw(seat Direction, t time.Time) (Tile, []Tile, error) {
	if r.CurrentTurn != seat {
		return "", nil, errors.New("wrong turn")
	}
	if r.CurrentAction != ActionDraw {
		return "", nil, errors.New("wrong action")
	}
	if len(r.Wall) < MinTilesLeft {
		return "", nil, errors.New("no draws left")
	}
	if t.Before(r.LastDiscardTime.Add(r.PongDuration)) {
		return "", nil, errors.New("waiting for other players to pong")
	}
	var draw Tile
	flowers := make([]Tile, 0)
	draw, r.Wall = drawFront(r.Wall)
	for isFlower(draw) {
		flowers = append(flowers, draw)
		r.Hands[seat].Flowers = append(r.Hands[seat].Flowers, draw)
		draw, r.Wall = drawBack(r.Wall)
	}
	r.Hands[seat].Concealed = append(r.Hands[seat].Concealed, draw)
	r.CurrentAction = ActionDiscard
	r.Events = append(r.Events, Event{
		Type:  EventDraw,
		Seat:  seat,
		Time:  t,
		Tiles: flowers,
	})
	return draw, flowers, nil
}

func indexOfPeng(revealed [][]Tile, tile Tile) int {
	for i, meld := range revealed {
		if len(meld) == 3 && meld[0] == tile && meld[1] == tile && meld[2] == tile {
			return i
		}
	}
	return -1
}

func (r *Round) lastDiscard() Tile {
	if len(r.Discards) > 0 {
		return r.Discards[len(r.Discards)-1]
	}
	return ""
}

func (r *Round) drawFlower(seat Direction) (Tile, []Tile) {
	flowers := make([]Tile, 0)
	for {
		var draw Tile
		draw, r.Wall = drawBack(r.Wall)
		if !isFlower(draw) {
			r.Hands[seat].Concealed = append(r.Hands[seat].Concealed, draw)
			return draw, flowers
		}
		r.Hands[seat].Flowers = append(r.Hands[seat].Flowers, draw)
		flowers = append(flowers, draw)
	}
}

func (r *Round) Kong(seat Direction, tile Tile) (Tile, []Tile, error) {
	if r.CurrentAction == ActionDraw && seat != r.PreviousTurn() && countTiles(r.Hands[seat].Concealed, tile) == 3 && r.lastDiscard() == tile {
		r.Discards = r.Discards[:len(r.Discards)-1]
		r.Hands[seat].Concealed, _ = removeTiles(r.Hands[seat].Concealed, tile, 3)
		r.Hands[seat].Revealed = append(r.Hands[seat].Revealed, []Tile{tile, tile, tile, tile})
		drawn, flowers := r.drawFlower(seat)
		r.CurrentAction = ActionDiscard
		r.CurrentTurn = seat
		r.Events = append(r.Events, Event{
			Type:  EventGang,
			Seat:  seat,
			Tiles: append([]Tile{tile}, flowers...),
		})
		return drawn, flowers, nil
	}
	if r.CurrentTurn == seat && r.CurrentAction == ActionDiscard {
		if i := indexOfPeng(r.Hands[seat].Revealed, tile); i != -1 && contains(r.Hands[seat].Concealed, tile) {
			r.Hands[seat].Concealed, _ = removeTile(r.Hands[seat].Concealed, tile)
			r.Hands[seat].Revealed[i] = append(r.Hands[seat].Revealed[i], tile)
			drawn, flowers := r.drawFlower(seat)
			r.Events = append(r.Events, Event{
				Type:  EventGang,
				Seat:  seat,
				Tiles: append([]Tile{tile}, flowers...),
			})
			return drawn, flowers, nil
		}
		if countTiles(r.Hands[seat].Concealed, tile) >= 4 {
			r.Hands[seat].Concealed, _ = removeTiles(r.Hands[seat].Concealed, tile, 4)
			r.Hands[seat].Revealed = append(r.Hands[seat].Revealed, []Tile{tile, tile, tile, tile})
			drawn, flowers := r.drawFlower(seat)
			r.Events = append(r.Events, Event{
				Type:  EventGang,
				Seat:  seat,
				Tiles: append([]Tile{tile}, flowers...),
			})
			return drawn, flowers, nil
		}
	}
	return "", nil, errors.New("not allowed")
}

func isChow(tiles []Tile) bool {
	return len(tiles) == 3 && validSequence([3]Tile{tiles[0], tiles[1], tiles[2]})
}

func isPeng(tiles []Tile) bool {
	return len(tiles) == 3 && tiles[0] == tiles[1] && tiles[1] == tiles[2]
}

func isKong(tiles []Tile) bool {
	return len(tiles) == 4 && tiles[0] == tiles[1] && tiles[1] == tiles[2] && tiles[2] == tiles[3]
}

func isEyes(tiles []Tile) bool {
	return len(tiles) == 2 && tiles[0] == tiles[1]
}

func checkWin(revealed [][]Tile, tiles []Tile, melds [][]Tile) bool {
	availableTiles := map[Tile]int{}
	for _, tile := range tiles {
		availableTiles[tile]++
	}
	hasEyes := false
	for _, meld := range melds {
		// check meld is valid
		switch {
		case isChow(meld):
		case isPeng(meld):
		case isEyes(meld):
			if hasEyes {
				// should only have one set of eyes
				return false
			} else {
				hasEyes = true
			}
		default:
			return false
		}
		// check player actually has the tiles
		for _, tile := range meld {
			if count := availableTiles[tile]; count > 0 {
				availableTiles[tile]--
			} else {
				return false
			}
		}
	}
	// length of revealed + melds should be 5 (4 sets of 3 + 1 set of eyes)
	if len(revealed)+len(melds) != 5 {
		return false
	}
	return true
}

func (r *Round) Win(seat Direction, melds [][]Tile) error {
	if r.CurrentAction == ActionDraw && seat != r.PreviousTurn() {
		if !checkWin(r.Hands[seat].Revealed, append(r.Hands[seat].Concealed, r.lastDiscard()), melds) {
			return errors.New("not allowed")
		}
		r.Discards = r.Discards[:len(r.Discards)-1]
		r.Hands[seat].Revealed = append(r.Hands[seat].Revealed, melds...)
		r.Hands[seat].Concealed = []Tile{}
		r.CurrentAction = ActionGameOver
		r.CurrentTurn = seat
		return nil
	}
	if seat == r.CurrentTurn && r.CurrentAction == ActionDiscard {
		if !checkWin(r.Hands[seat].Revealed, r.Hands[seat].Concealed, melds) {
			return errors.New("not allowed")
		}
		r.Hands[seat].Revealed = append(r.Hands[seat].Revealed, melds...)
		r.Hands[seat].Concealed = []Tile{}
		r.CurrentAction = ActionGameOver
		return nil
	}
	return errors.New("not allowed")
}

// EndGame ends the game in a draw.
func (r *Round) EndGame(seat Direction) error {
	if r.CurrentTurn == seat && len(r.Wall) < MinTilesLeft && r.CurrentAction == ActionDiscard {
		r.CurrentAction = ActionGameOver
		r.CurrentTurn = -1
		return nil
	}
	return errors.New("not allowed")
}

// durationMillis is a custom type to facilitate marshalling a time.Duration to JSON.
type durationMillis time.Duration

func (d durationMillis) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(int(time.Duration(d).Milliseconds()))), nil
}

// RoundView represents a view of a round with non-public information removed.
type RoundView struct {
	DrawsLeft       int            `json:"draws_left"`
	Discards        []Tile         `json:"discards"`
	Hands           [4]Hand        `json:"hands"`
	CurrentTurn     Direction      `json:"current_turn"`
	CurrentAction   Action         `json:"current_action"`
	PongDuration    durationMillis `json:"pong_duration"`
	LastDiscardTime int64          `json:"last_discard_time"`
	Events          []Event        `json:"events"`
}

// ViewFromSeat returns a RoundView from a specific seat.
func (r *Round) ViewFromSeat(seat Direction) RoundView {
	var hands [4]Hand
	for i, hand := range r.Hands {
		if Direction(i) == seat {
			hands[i] = hand
		} else {
			hands[i] = hand.Masked()
		}
	}
	return RoundView{
		DrawsLeft:       len(r.Wall) - MinTilesLeft + 1,
		Discards:        r.Discards,
		Hands:           hands,
		CurrentTurn:     r.CurrentTurn,
		CurrentAction:   r.CurrentAction,
		PongDuration:    durationMillis(r.PongDuration),
		LastDiscardTime: r.LastDiscardTime.UnixNano() / 1e6,
		Events:          r.Events,
	}
}

func NewRound(seed int64) *Round {
	r := rand.New(rand.NewSource(seed))
	wall := newWall(r)
	hands, wall := distributeTiles(wall)
	return &Round{
		Discards:      []Tile{},
		Wall:          wall,
		Hands:         hands,
		CurrentTurn:   DirectionEast,
		CurrentAction: ActionDiscard,
		Events:        []Event{},
	}
}
