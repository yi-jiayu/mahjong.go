package mahjong

import (
	"encoding/json"
	"errors"
	"math/rand"
	"sort"
)

const (
	TileCat          = "01猫"
	TileRat          = "02老鼠"
	TileRooster      = "03公鸡"
	TileCentipede    = "04蜈蚣"
	TileGentlemen1   = "05梅"
	TileGentlemen2   = "06兰"
	TileGentlemen3   = "07菊"
	TileGentlemen4   = "08竹"
	TileSeasons1     = "09春"
	TileSeasons2     = "10夏"
	TileSeasons3     = "11秋"
	TileSeasons4     = "12冬"
	TileDots1        = "13一筒"
	TileDots2        = "14二筒"
	TileDots3        = "15三筒"
	TileDots4        = "16四筒"
	TileDots5        = "17五筒"
	TileDots6        = "18六筒"
	TileDots7        = "19七筒"
	TileDots8        = "20八筒"
	TileDots9        = "21九筒"
	TileBamboo1      = "22一索"
	TileBamboo2      = "23二索"
	TileBamboo3      = "24三索"
	TileBamboo4      = "25四索"
	TileBamboo5      = "26五索"
	TileBamboo6      = "27六索"
	TileBamboo7      = "28七索"
	TileBamboo8      = "29八索"
	TileBamboo9      = "30九索"
	TileCharacters1  = "31一万"
	TileCharacters2  = "32二万"
	TileCharacters3  = "33三万"
	TileCharacters4  = "34四万"
	TileCharacters5  = "35五万"
	TileCharacters6  = "36六万"
	TileCharacters7  = "37七万"
	TileCharacters8  = "38八万"
	TileCharacters9  = "39九万"
	TileWindsEast    = "40东风"
	TileWindsSouth   = "41南风"
	TileWindsWest    = "42西风"
	TileWindsNorth   = "43北风"
	TileDragonsRed   = "44红中"
	TileDragonsGreen = "45青发"
	TileDragonsWhite = "46白板"
)

const (
	DirectionEast = iota
	DirectionSouth
	DirectionWest
	DirectionNorth
)

const (
	ActionDraw     = "draw"
	ActionDiscard  = "discard"
	ActionGameOver = "game over"
)

var (
	FlowerTiles = []string{"01猫", "02老鼠", "03公鸡", "04蜈蚣", "05梅", "06兰", "07菊", "08竹", "09春", "10夏", "11秋", "12冬"}
	NormalTiles = []string{"13一筒", "14二筒", "15三筒", "16四筒", "17五筒", "18六筒", "19七筒", "20八筒", "21九筒", "22一索", "23二索", "24三索", "25四索", "26五索", "27六索", "28七索", "29八索", "30九索", "31一万", "32二万", "33三万", "34四万", "35五万", "36六万", "37七万", "38八万", "39九万", "40东风", "41南风", "42西风", "43北风", "44红中", "45青发", "46白板"}
)

var validSequences = [][3]string{
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

type Hand struct {
	Flowers   []string
	Revealed  []string
	Concealed []string
}

func (h Hand) MarshalJSON() ([]byte, error) {
	masked := make([]string, len(h.Concealed))
	return json.Marshal(struct {
		Flowers   []string `json:"flowers"`
		Revealed  []string `json:"revealed"`
		Concealed []string `json:"concealed"`
	}{
		Flowers:   h.Flowers,
		Revealed:  h.Revealed,
		Concealed: masked,
	})
}

type Round struct {
	Wall           []string
	Discards       []string
	Hands          [4]Hand
	SequenceNumber int
	CurrentTurn    int
	CurrentAction  string
}

func (r *Round) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		DrawsLeft      int      `json:"draws_left"`
		Discards       []string `json:"discards"`
		Hands          [4]Hand  `json:"hands"`
		SequenceNumber int      `json:"sequence_number"`
		CurrentTurn    int      `json:"current_turn"`
		CurrentAction  string   `json:"current_action"`
	}{
		DrawsLeft:      len(r.Wall),
		Discards:       r.Discards,
		Hands:          r.Hands,
		SequenceNumber: r.SequenceNumber,
		CurrentTurn:    r.CurrentTurn,
		CurrentAction:  r.CurrentAction,
	})
}

func newWall(r *rand.Rand) []string {
	var wall []string
	wall = append(wall, FlowerTiles...)
	for _, tile := range NormalTiles {
		wall = append(wall, tile, tile, tile, tile)
	}
	r.Shuffle(len(wall), func(i, j int) {
		wall[i], wall[j] = wall[j], wall[i]
	})
	return wall
}

func drawFront(wall []string) (string, []string) {
	drawn := wall[0]
	wall = wall[1:]
	return drawn, wall
}

func drawFrontN(wall []string, n int) ([]string, []string) {
	drawn := wall[:n]
	wall = wall[n:]
	return drawn, wall
}

func drawBack(wall []string) (string, []string) {
	drawn := wall[len(wall)-1]
	wall = wall[:len(wall)-1]
	return drawn, wall
}

func contains(tiles []string, tile string) bool {
	for _, t := range tiles {
		if t == tile {
			return true
		}
	}
	return false
}

func isFlower(tile string) bool {
	return contains(FlowerTiles, tile)
}

func distributeTiles(wall []string, dealer int) ([4]Hand, []string) {
	hands := [4]Hand{
		{
			Flowers:   []string{},
			Revealed:  []string{},
			Concealed: []string{},
		},
		{
			Flowers:   []string{},
			Revealed:  []string{},
			Concealed: []string{},
		},
		{
			Flowers:   []string{},
			Revealed:  []string{},
			Concealed: []string{},
		},
		{
			Flowers:   []string{},
			Revealed:  []string{},
			Concealed: []string{},
		},
	}
	order := []int{dealer, (dealer + 1) % 4, (dealer + 2) % 4, (dealer + 3) % 4}
	// draw 4 tiles 3 times
	for i := 0; i < 3; i++ {
		var draws []string
		for _, seat := range order {
			draws, wall = drawFrontN(wall, 4)
			hands[seat].Concealed = append(hands[seat].Concealed, draws...)
		}
	}
	// draw one tile
	var draw string
	for _, seat := range order {
		draw, wall = drawFront(wall)
		hands[seat].Concealed = append(hands[seat].Concealed, draw)
	}
	// dealer draws one extra tile
	draw, wall = drawFront(wall)
	hands[dealer].Concealed = append(hands[dealer].Concealed, draw)
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

func removeTiles(tiles []string, tile string, count int) ([]string, int) {
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

func removeTile(tiles []string, tile string) ([]string, bool) {
	tiles, removedCount := removeTiles(tiles, tile, 1)
	return tiles, removedCount > 0
}

func (r *Round) Discard(seat int, tile string) error {
	if r.CurrentTurn != seat {
		return errors.New("not your turn")
	}
	if r.CurrentAction != ActionDiscard {
		return errors.New("not time to discard")
	}
	remaining, ok := removeTile(r.Hands[seat].Concealed, tile)
	if !ok {
		return errors.New("no such tile")
	}
	r.Hands[seat].Concealed = remaining
	r.Discards = append(r.Discards, tile)
	r.CurrentTurn = (seat + 1) % 4
	r.CurrentAction = ActionDraw
	r.SequenceNumber++
	return nil
}

func validSequence(seq [3]string) bool {
	sort.Strings(seq[:])
	for _, valid := range validSequences {
		if seq == valid {
			return true
		}
	}
	return false
}

func (r *Round) Chow(seat int, tile1, tile2 string) error {
	if r.CurrentTurn != seat {
		return errors.New("not your turn")
	}
	if r.CurrentAction != ActionDraw {
		return errors.New("not time to chow")
	}
	if !contains(r.Hands[seat].Concealed, tile1) || !contains(r.Hands[seat].Concealed, tile2) {
		return errors.New("no such tile")
	}
	seq := [3]string{tile1, tile2, r.Discards[len(r.Discards)-1]}
	if !validSequence(seq) {
		return errors.New("invalid sequence")
	}
	r.Hands[seat].Concealed, _ = removeTile(r.Hands[seat].Concealed, tile1)
	r.Hands[seat].Concealed, _ = removeTile(r.Hands[seat].Concealed, tile2)
	r.Hands[seat].Revealed = append(r.Hands[seat].Revealed, seq[:]...)
	r.Discards = r.Discards[:len(r.Discards)-1]
	r.CurrentAction = ActionDiscard
	r.SequenceNumber++
	return nil
}

func (r *Round) PreviousTurn() int {
	return (r.CurrentTurn + 3) % 4
}

func countTiles(tiles []string, tile string) int {
	count := 0
	for _, t := range tiles {
		if t == tile {
			count++
		}
	}
	return count
}

func (r *Round) Peng(seat int, tile string) error {
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
	r.Hands[seat].Revealed = append(r.Hands[seat].Revealed, tile, tile, tile)
	r.Discards = r.Discards[:len(r.Discards)-1]
	r.CurrentAction = ActionDiscard
	r.CurrentTurn = seat
	r.SequenceNumber++
	return nil
}

func (r *Round) Draw(seat int) error {
	if r.CurrentTurn != seat {
		return errors.New("wrong turn")
	}
	if r.CurrentAction != ActionDraw {
		return errors.New("wrong action")
	}
	var draw string
	draw, r.Wall = drawFront(r.Wall)
	for isFlower(draw) {
		r.Hands[seat].Flowers = append(r.Hands[seat].Flowers, draw)
		draw, r.Wall = drawBack(r.Wall)
	}
	r.Hands[seat].Concealed = append(r.Hands[seat].Concealed, draw)
	r.CurrentAction = ActionDiscard
	r.SequenceNumber++
	return nil
}

func NewRound(seed int64, dealer int) *Round {
	r := rand.New(rand.NewSource(seed))
	wall := newWall(r)
	hands, wall := distributeTiles(wall, dealer)
	return &Round{
		Discards:      []string{},
		Wall:          wall,
		Hands:         hands,
		CurrentTurn:   dealer,
		CurrentAction: ActionDiscard,
	}
}
