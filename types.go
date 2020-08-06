package mahjong

import (
	"errors"
	"math/rand"
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

type Hand struct {
	Flowers   []string
	Revealed  []string
	Concealed []string
}

type Round struct {
	Wall           []string
	Discards       []string
	Hands          []Hand
	PrevailingWind int
	CurrentTurn    int
	CurrentAction  string
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

func isFlower(tile string) bool {
	for _, flower := range FlowerTiles {
		if tile == flower {
			return true
		}
	}
	return false
}

func contains(tiles []string, tile string) bool {
	for _, t := range tiles {
		if t == tile {
			return true
		}
	}
	return false
}

func distributeTiles(wall []string, dealer int) ([]Hand, []string) {
	hands := make([]Hand, 4)
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

func (r *Round) Discard(seat int, tile string) error {
	if r.CurrentTurn != seat {
		return errors.New("not your turn")
	}
	if r.CurrentAction != ActionDiscard {
		return errors.New("not time to discard")
	}
	if !contains(r.Hands[seat].Concealed, tile) {
		return errors.New("no such tile")
	}
	i := 0
	for _, t := range r.Hands[seat].Concealed {
		if t == tile {
			r.Discards = append(r.Discards, t)
			tile = ""
		} else {
			r.Hands[seat].Concealed[i] = t
			i++
		}
	}
	r.Hands[seat].Concealed = r.Hands[seat].Concealed[i:]
	r.CurrentTurn = (seat + 1) % 4
	r.CurrentAction = ActionDiscard
	return nil
}

func NewRound(seed int64, wind int, dealer int) *Round {
	r := rand.New(rand.NewSource(seed))
	wall := newWall(r)
	hands, wall := distributeTiles(wall, dealer)
	return &Round{
		Wall:           wall,
		Hands:          hands,
		PrevailingWind: wind,
		CurrentTurn:    dealer,
		CurrentAction:  ActionDiscard,
	}
}
