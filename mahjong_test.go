package mahjong

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newWall(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	got := newWall(r)
	want := []Tile{"38八万", "35五万", "27六索", "44红中", "22一索", "34四万", "35五万", "20八筒", "37七万", "13一筒", "43北风", "26五索", "21九筒", "25四索", "42西风", "17五筒", "38八万", "36六万", "16四筒", "43北风", "20八筒", "22一索", "37七万", "25四索", "42西风", "30九索", "19七筒", "06兰", "27六索", "07菊", "40东风", "32二万", "29八索", "36六万", "34四万", "46白板", "32二万", "15三筒", "17五筒", "37七万", "42西风", "14二筒", "43北风", "20八筒", "28七索", "45青发", "17五筒", "36六万", "34四万", "14二筒", "12冬", "46白板", "22一索", "40东风", "37七万", "28七索", "29八索", "16四筒", "39九万", "13一筒", "24三索", "01猫", "27六索", "40东风", "41南风", "34四万", "24三索", "31一万", "31一万", "25四索", "13一筒", "26五索", "15三筒", "14二筒", "18六筒", "24三索", "11秋", "19七筒", "45青发", "41南风", "44红中", "39九万", "27六索", "26五索", "10夏", "15三筒", "21九筒", "36六万", "41南风", "33三万", "29八索", "23二索", "28七索", "04蜈蚣", "32二万", "38八万", "29八索", "05梅", "39九万", "21九筒", "46白板", "33三万", "09春", "32二万", "25四索", "30九索", "39九万", "23二索", "02老鼠", "24三索", "44红中", "28七索", "45青发", "18六筒", "31一万", "14二筒", "43北风", "13一筒", "45青发", "30九索", "18六筒", "22一索", "31一万", "16四筒", "17五筒", "26五索", "23二索", "21九筒", "35五万", "42西风", "03公鸡", "35五万", "18六筒", "30九索", "46白板", "38八万", "40东风", "19七筒", "15三筒", "41南风", "33三万", "16四筒", "20八筒", "23二索", "08竹", "33三万", "19七筒", "44红中"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, wanted %v", got, want)
	}
}

func Test_distributeTiles(t *testing.T) {
	t.Run("no flowers", func(t *testing.T) {
		wall := []Tile{
			// first 3 draws of 4 tiles
			"38八万", "35五万", "27六索", "44红中",
			"22一索", "34四万", "35五万", "20八筒",
			"37七万", "13一筒", "43北风", "26五索",
			"21九筒", "25四索", "42西风", "17五筒",

			"38八万", "36六万", "16四筒", "43北风",
			"20八筒", "22一索", "37七万", "25四索",
			"42西风", "30九索", "19七筒", "40东风",
			"27六索", "37七万", "40东风", "32二万",

			"29八索", "36六万", "34四万", "46白板",
			"32二万", "15三筒", "17五筒", "37七万",
			"42西风", "14二筒", "43北风", "20八筒",
			"28七索", "45青发", "17五筒", "36六万",

			// single tile draws
			"34四万",
			"14二筒",
			"28七索",
			"46白板",

			"22一索", // dealer extra tile

			// rest of wall
			"06兰", "07菊", "12冬", "29八索", "16四筒", "39九万", "13一筒", "24三索", "01猫", "27六索", "40东风", "41南风", "34四万", "24三索", "31一万", "31一万", "25四索", "13一筒", "26五索", "15三筒", "14二筒", "18六筒", "24三索", "11秋", "19七筒", "45青发", "41南风", "44红中", "39九万", "27六索", "26五索", "10夏", "15三筒", "21九筒", "36六万", "41南风", "33三万", "29八索", "23二索", "28七索", "04蜈蚣", "32二万", "38八万", "29八索", "05梅", "39九万", "21九筒", "46白板", "33三万", "09春", "32二万", "25四索", "30九索", "39九万", "23二索", "02老鼠", "24三索", "44红中", "28七索", "45青发", "18六筒", "31一万", "14二筒", "43北风", "13一筒", "45青发", "30九索", "18六筒", "22一索", "31一万", "16四筒", "17五筒", "26五索", "23二索", "21九筒", "35五万", "42西风", "03公鸡", "35五万", "18六筒", "30九索", "46白板", "38八万", "40东风", "19七筒", "15三筒", "41南风", "33三万", "16四筒", "20八筒", "23二索", "08竹", "33三万", "19七筒", "44红中",
		}
		hands, wall := distributeTiles(wall)
		assert.Equal(t,
			[]Tile{"38八万", "35五万", "27六索", "44红中", "38八万", "36六万", "16四筒", "43北风", "29八索", "36六万", "34四万", "46白板", "34四万", "22一索"},
			hands[DirectionEast].Concealed)
		assert.Equal(t,
			[]Tile{"22一索", "34四万", "35五万", "20八筒", "20八筒", "22一索", "37七万", "25四索", "32二万", "15三筒", "17五筒", "37七万", "14二筒"},
			hands[DirectionSouth].Concealed)
		assert.Equal(t,
			[]Tile{"37七万", "13一筒", "43北风", "26五索", "42西风", "30九索", "19七筒", "40东风", "42西风", "14二筒", "43北风", "20八筒", "28七索"},
			hands[DirectionWest].Concealed)
		assert.Equal(t,
			[]Tile{"21九筒", "25四索", "42西风", "17五筒", "27六索", "37七万", "40东风", "32二万", "28七索", "45青发", "17五筒", "36六万", "46白板"},
			hands[DirectionNorth].Concealed)
		assert.Equal(t,
			[]Tile{"06兰", "07菊", "12冬", "29八索", "16四筒", "39九万", "13一筒", "24三索", "01猫", "27六索", "40东风", "41南风", "34四万", "24三索", "31一万", "31一万", "25四索", "13一筒", "26五索", "15三筒", "14二筒", "18六筒", "24三索", "11秋", "19七筒", "45青发", "41南风", "44红中", "39九万", "27六索", "26五索", "10夏", "15三筒", "21九筒", "36六万", "41南风", "33三万", "29八索", "23二索", "28七索", "04蜈蚣", "32二万", "38八万", "29八索", "05梅", "39九万", "21九筒", "46白板", "33三万", "09春", "32二万", "25四索", "30九索", "39九万", "23二索", "02老鼠", "24三索", "44红中", "28七索", "45青发", "18六筒", "31一万", "14二筒", "43北风", "13一筒", "45青发", "30九索", "18六筒", "22一索", "31一万", "16四筒", "17五筒", "26五索", "23二索", "21九筒", "35五万", "42西风", "03公鸡", "35五万", "18六筒", "30九索", "46白板", "38八万", "40东风", "19七筒", "15三筒", "41南风", "33三万", "16四筒", "20八筒", "23二索", "08竹", "33三万", "19七筒", "44红中"},
			wall)
	})
	t.Run("replace flowers", func(t *testing.T) {
		wall := []Tile{
			"38八万", "35五万", "27六索", "44红中",
			"22一索", "34四万", "35五万", "20八筒",
			"37七万", "13一筒", "43北风", "26五索",
			"21九筒", "25四索", "42西风", "17五筒",

			"38八万", "36六万", "16四筒", "43北风",
			"20八筒", "22一索", "37七万", "25四索",
			"42西风", "30九索", "19七筒", "06兰", // west draws a flower
			"27六索", "07菊", "40东风", "32二万", // north draws a flower

			"29八索", "36六万", "34四万", "46白板",
			"32二万", "15三筒", "17五筒", "37七万",
			"42西风", "14二筒", "43北风", "20八筒",
			"28七索", "45青发", "17五筒", "36六万",

			"34四万",
			"14二筒",
			"12冬", // west draws another flower
			"46白板",

			"22一索",

			"40东风", "37七万", "28七索", "29八索", "16四筒", "39九万", "13一筒", "24三索", "01猫", "27六索", "40东风", "41南风", "34四万", "24三索", "31一万", "31一万", "25四索", "13一筒", "26五索", "15三筒", "14二筒", "18六筒", "24三索", "11秋", "19七筒", "45青发", "41南风", "44红中", "39九万", "27六索", "26五索", "10夏", "15三筒", "21九筒", "36六万", "41南风", "33三万", "29八索", "23二索", "28七索", "04蜈蚣", "32二万", "38八万", "29八索", "05梅", "39九万", "21九筒", "46白板", "33三万", "09春", "32二万", "25四索", "30九索", "39九万", "23二索", "02老鼠", "24三索", "44红中", "28七索", "45青发", "18六筒", "31一万", "14二筒", "43北风", "13一筒", "45青发", "30九索", "18六筒", "22一索", "31一万", "16四筒", "17五筒", "26五索", "23二索", "21九筒", "35五万", "42西风", "03公鸡", "35五万", "18六筒", "30九索", "46白板", "38八万", "40东风", "19七筒", "15三筒", "41南风", "33三万", "16四筒", "20八筒", "23二索", "08竹",

			"33三万",         // north replaces one tile
			"19七筒", "44红中", // west replaces two tiles
		}
		hands, wall := distributeTiles(wall)
		assert.Equal(t,
			[]Tile{"38八万", "35五万", "27六索", "44红中", "38八万", "36六万", "16四筒", "43北风", "29八索", "36六万", "34四万", "46白板", "34四万", "22一索"},
			hands[DirectionEast].Concealed)
		assert.Equal(t,
			[]Tile{"22一索", "34四万", "35五万", "20八筒", "20八筒", "22一索", "37七万", "25四索", "32二万", "15三筒", "17五筒", "37七万", "14二筒"},
			hands[DirectionSouth].Concealed)
		assert.ElementsMatch(t,
			[]Tile{
				"37七万", "13一筒", "43北风", "26五索",
				"42西风", "30九索", "19七筒",
				"42西风", "14二筒", "43北风", "20八筒",
				"19七筒", "44红中", // replaced tiles
			},
			hands[DirectionWest].Concealed)
		assert.ElementsMatch(t, []Tile{"06兰", "12冬"}, hands[DirectionWest].Flowers)
		assert.ElementsMatch(t,
			[]Tile{
				"21九筒", "25四索", "42西风", "17五筒",
				"27六索", "40东风", "32二万",
				"28七索", "45青发", "17五筒", "36六万",
				"46白板",
				"33三万", // replaced tile
			},
			hands[DirectionNorth].Concealed)
		assert.ElementsMatch(t, []Tile{"07菊"}, hands[DirectionNorth].Flowers)
		assert.Equal(t,
			[]Tile{"40东风", "37七万", "28七索", "29八索", "16四筒", "39九万", "13一筒", "24三索", "01猫", "27六索", "40东风", "41南风", "34四万", "24三索", "31一万", "31一万", "25四索", "13一筒", "26五索", "15三筒", "14二筒", "18六筒", "24三索", "11秋", "19七筒", "45青发", "41南风", "44红中", "39九万", "27六索", "26五索", "10夏", "15三筒", "21九筒", "36六万", "41南风", "33三万", "29八索", "23二索", "28七索", "04蜈蚣", "32二万", "38八万", "29八索", "05梅", "39九万", "21九筒", "46白板", "33三万", "09春", "32二万", "25四索", "30九索", "39九万", "23二索", "02老鼠", "24三索", "44红中", "28七索", "45青发", "18六筒", "31一万", "14二筒", "43北风", "13一筒", "45青发", "30九索", "18六筒", "22一索", "31一万", "16四筒", "17五筒", "26五索", "23二索", "21九筒", "35五万", "42西风", "03公鸡", "35五万", "18六筒", "30九索", "46白板", "38八万", "40东风", "19七筒", "15三筒", "41南风", "33三万", "16四筒", "20八筒", "23二索", "08竹"},
			wall)
	})
	t.Run("replacing flowers again", func(t *testing.T) {
		wall := []Tile{
			"38八万", "35五万", "27六索", "44红中",
			"22一索", "34四万", "35五万", "20八筒",
			"37七万", "13一筒", "43北风", "26五索",
			"21九筒", "25四索", "42西风", "17五筒",

			"38八万", "36六万", "16四筒", "43北风",
			"20八筒", "22一索", "37七万", "25四索",
			"42西风", "30九索", "19七筒", "06兰", // west draws a flower
			"27六索", "07菊", "40东风", "32二万", // north draws a flower

			"29八索", "36六万", "34四万", "46白板",
			"32二万", "15三筒", "17五筒", "37七万",
			"42西风", "14二筒", "43北风", "20八筒",
			"28七索", "45青发", "17五筒", "36六万",

			"34四万",
			"14二筒",
			"12冬", // west draws another flower
			"46白板",

			"22一索",

			"40东风", "37七万", "28七索", "29八索", "16四筒", "39九万", "13一筒", "24三索", "44红中", "27六索", "40东风", "41南风", "34四万", "24三索", "31一万", "31一万", "25四索", "13一筒", "26五索", "15三筒", "14二筒", "18六筒", "24三索", "11秋", "19七筒", "45青发", "41南风", "44红中", "39九万", "27六索", "26五索", "10夏", "15三筒", "21九筒", "36六万", "41南风", "33三万", "29八索", "23二索", "28七索", "04蜈蚣", "32二万", "38八万", "29八索", "05梅", "39九万", "21九筒", "46白板", "33三万", "09春", "32二万", "25四索", "30九索", "39九万", "23二索", "02老鼠", "24三索", "44红中", "28七索", "45青发", "18六筒", "31一万", "14二筒", "43北风", "13一筒", "45青发", "30九索", "18六筒", "22一索", "31一万", "16四筒", "17五筒", "26五索", "23二索", "21九筒", "35五万", "42西风", "03公鸡", "35五万", "18六筒", "30九索", "46白板", "38八万", "40东风", "19七筒", "15三筒", "41南风", "33三万", "16四筒", "20八筒",

			"23二索",        // west replaces the fourth flower
			"08竹",         // west replaces the third flower and gets a fourth flower
			"33三万",        // north replaces one tile
			"19七筒", "01猫", // west replaces two tiles and gets a third flower
		}
		hands, wall := distributeTiles(wall)
		assert.Equal(t,
			[]Tile{"38八万", "35五万", "27六索", "44红中", "38八万", "36六万", "16四筒", "43北风", "29八索", "36六万", "34四万", "46白板", "34四万", "22一索"},
			hands[DirectionEast].Concealed)
		assert.Equal(t,
			[]Tile{"22一索", "34四万", "35五万", "20八筒", "20八筒", "22一索", "37七万", "25四索", "32二万", "15三筒", "17五筒", "37七万", "14二筒"},
			hands[DirectionSouth].Concealed)
		assert.ElementsMatch(t,
			[]Tile{
				"37七万", "13一筒", "43北风", "26五索",
				"42西风", "30九索", "19七筒",
				"42西风", "14二筒", "43北风", "20八筒",
				"19七筒", "23二索", // replaced tiles
			},
			hands[DirectionWest].Concealed)
		assert.ElementsMatch(t, []Tile{"06兰", "12冬", "01猫", "08竹"}, hands[DirectionWest].Flowers)
		assert.ElementsMatch(t,
			[]Tile{
				"21九筒", "25四索", "42西风", "17五筒",
				"27六索", "40东风", "32二万",
				"28七索", "45青发", "17五筒", "36六万",
				"46白板",
				"33三万", // replaced tile
			},
			hands[DirectionNorth].Concealed)
		assert.ElementsMatch(t, []Tile{"07菊"}, hands[DirectionNorth].Flowers)
		assert.Equal(t,
			[]Tile{"40东风", "37七万", "28七索", "29八索", "16四筒", "39九万", "13一筒", "24三索", "44红中", "27六索", "40东风", "41南风", "34四万", "24三索", "31一万", "31一万", "25四索", "13一筒", "26五索", "15三筒", "14二筒", "18六筒", "24三索", "11秋", "19七筒", "45青发", "41南风", "44红中", "39九万", "27六索", "26五索", "10夏", "15三筒", "21九筒", "36六万", "41南风", "33三万", "29八索", "23二索", "28七索", "04蜈蚣", "32二万", "38八万", "29八索", "05梅", "39九万", "21九筒", "46白板", "33三万", "09春", "32二万", "25四索", "30九索", "39九万", "23二索", "02老鼠", "24三索", "44红中", "28七索", "45青发", "18六筒", "31一万", "14二筒", "43北风", "13一筒", "45青发", "30九索", "18六筒", "22一索", "31一万", "16四筒", "17五筒", "26五索", "23二索", "21九筒", "35五万", "42西风", "03公鸡", "35五万", "18六筒", "30九索", "46白板", "38八万", "40东风", "19七筒", "15三筒", "41南风", "33三万", "16四筒", "20八筒"},
			wall)
	})
}

func TestRound_Discard(t *testing.T) {
	t.Run("wrong turn", func(t *testing.T) {
		round := &Round{CurrentTurn: DirectionEast}
		err := round.Discard(DirectionSouth, "")
		assert.Error(t, err)
	})
	t.Run("wrong action", func(t *testing.T) {
		round := &Round{CurrentAction: ActionDraw}
		err := round.Discard(DirectionEast, "")
		assert.Error(t, err)
	})
	t.Run("no such tile", func(t *testing.T) {
		round := &Round{
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDiscard,
			Hands:         [4]Hand{{}},
		}
		err := round.Discard(DirectionEast, "")
		assert.Error(t, err)
	})
	t.Run("cannot discard last tile", func(t *testing.T) {
		round := &Round{
			Wall: []Tile{
				"38八万", "35五万", "27六索", "44红中",
				"22一索", "34四万", "35五万", "20八筒",
				"37七万", "13一筒", "43北风", "26五索",
				"21九筒", "25四索", "42西风",
			},
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDiscard,
		}
		err := round.Discard(DirectionEast, "")
		assert.EqualError(t, err, "cannot discard on last round")
	})
	t.Run("success", func(t *testing.T) {
		round := &Round{
			Wall: []Tile{"38八万", "35五万", "27六索", "44红中",
				"22一索", "34四万", "35五万", "20八筒",
				"37七万", "13一筒", "43北风", "26五索",
				"21九筒", "25四索", "42西风", "17五筒"},
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDiscard,
			Hands:         [4]Hand{{Concealed: []Tile{TileWindsEast, TileWindsEast, TileWindsEast}}},
		}
		err := round.Discard(DirectionEast, TileWindsEast)
		assert.NoError(t, err)
		assert.Equal(t, []Tile{TileWindsEast}, round.Discards)
		assert.Equal(t, []Tile{TileWindsEast, TileWindsEast}, round.Hands[DirectionEast].Concealed)
		assert.Equal(t, DirectionSouth, round.CurrentTurn)
		assert.Equal(t, ActionDraw, round.CurrentAction)
	})
}

func TestRound_Chow(t *testing.T) {
	t.Run("wrong turn", func(t *testing.T) {
		round := &Round{CurrentTurn: DirectionEast}
		err := round.Chow(DirectionSouth, "", "")
		assert.Error(t, err)
	})
	t.Run("wrong action", func(t *testing.T) {
		round := &Round{CurrentAction: ActionDiscard}
		err := round.Chow(DirectionEast, "", "")
		assert.Error(t, err)
	})
	t.Run("no such tiles", func(t *testing.T) {
		round := &Round{
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
			Hands:         [4]Hand{{}},
		}
		err := round.Chow(DirectionEast, TileBamboo1, TileBamboo2)
		assert.Error(t, err)
	})
	t.Run("invalid sequence", func(t *testing.T) {
		round := &Round{
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
			Discards:      []Tile{TileBamboo4},
			Hands:         [4]Hand{{Concealed: []Tile{TileBamboo1, TileBamboo2}}},
		}
		err := round.Chow(DirectionEast, TileBamboo1, TileBamboo2)
		assert.Error(t, err)
	})
	t.Run("success", func(t *testing.T) {
		round := &Round{
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
			Discards:      []Tile{TileBamboo4, TileBamboo3},
			Hands:         [4]Hand{{Concealed: []Tile{TileWindsWest, TileBamboo1, TileBamboo2}}},
		}
		err := round.Chow(DirectionEast, TileBamboo1, TileBamboo2)
		assert.NoError(t, err)
		assert.Equal(t, []Tile{TileBamboo4}, round.Discards)
		assert.Equal(t, []Tile{TileWindsWest}, round.Hands[DirectionEast].Concealed)
		assert.Equal(t, [][]Tile{{TileBamboo1, TileBamboo2, TileBamboo3}}, round.Hands[DirectionEast].Revealed)
		assert.Equal(t, DirectionEast, round.CurrentTurn)
		assert.Equal(t, ActionDiscard, round.CurrentAction)
	})
}

func Test_validSequence(t *testing.T) {
	tests := []struct {
		Name     string
		Sequence [3]Tile
		Valid    bool
	}{
		{
			Name:     "in order",
			Sequence: [3]Tile{TileBamboo1, TileBamboo2, TileBamboo3},
			Valid:    true,
		},
		{
			Name:     "out of order",
			Sequence: [3]Tile{TileBamboo2, TileBamboo1, TileBamboo3},
			Valid:    true,
		},
		{
			Name:     "invalid",
			Sequence: [3]Tile{TileBamboo1, TileBamboo2, TileBamboo4},
			Valid:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := validSequence(tt.Sequence)
			assert.Equal(t, tt.Valid, got)
		})
	}
}

func TestRound_Peng(t *testing.T) {
	t.Run("wrong turn", func(t *testing.T) {
		round := &Round{CurrentTurn: DirectionEast}
		err := round.Peng(DirectionNorth, "")
		assert.EqualError(t, err, "wrong turn")
	})
	t.Run("wrong action", func(t *testing.T) {
		round := &Round{CurrentAction: ActionDiscard}
		err := round.Peng(DirectionEast, "")
		assert.EqualError(t, err, "wrong action")
	})
	t.Run("no such tile", func(t *testing.T) {
		round := &Round{
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
			Hands:         [4]Hand{{}},
		}
		err := round.Peng(DirectionEast, TileBamboo1)
		assert.EqualError(t, err, "not enough tiles")
	})
	t.Run("not enough tiles", func(t *testing.T) {
		round := &Round{
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
			Hands:         [4]Hand{{Concealed: []Tile{TileBamboo1}}},
		}
		err := round.Peng(DirectionEast, TileBamboo1)
		assert.EqualError(t, err, "not enough tiles")
	})
	t.Run("success", func(t *testing.T) {
		round := &Round{
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
			Discards:      []Tile{TileBamboo4, TileBamboo4},
			Hands:         [4]Hand{{}, {Concealed: []Tile{TileWindsWest, TileBamboo4, TileBamboo4}}},
		}
		err := round.Peng(DirectionSouth, TileBamboo4)
		assert.NoError(t, err)
		assert.Equal(t, []Tile{TileBamboo4}, round.Discards)
		assert.Equal(t, []Tile{TileWindsWest}, round.Hands[DirectionSouth].Concealed)
		assert.Equal(t, [][]Tile{{TileBamboo4, TileBamboo4, TileBamboo4}}, round.Hands[DirectionSouth].Revealed)
		assert.Equal(t, DirectionSouth, round.CurrentTurn)
		assert.Equal(t, ActionDiscard, round.CurrentAction)
	})
}

func TestRound_Draw(t *testing.T) {
	t.Run("wrong turn", func(t *testing.T) {
		round := &Round{CurrentTurn: DirectionEast}
		err := round.Draw(DirectionSouth)
		assert.EqualError(t, err, "wrong turn")
	})
	t.Run("wrong action", func(t *testing.T) {
		round := &Round{CurrentAction: ActionDiscard}
		err := round.Draw(DirectionEast)
		assert.EqualError(t, err, "wrong action")
	})
	t.Run("cannot draw when less than MinTilesLeft tiles in wall", func(t *testing.T) {
		round := &Round{
			Wall: []Tile{
				"38八万", "35五万", "27六索", "44红中",
				"22一索", "34四万", "35五万", "20八筒",
				"37七万", "13一筒", "43北风", "26五索",
				"21九筒", "25四索", "42西风",
			},
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
		}
		err := round.Draw(DirectionEast)
		assert.EqualError(t, err, "no draws left")
	})
	t.Run("success", func(t *testing.T) {
		round := &Round{
			Wall: []Tile{
				TileBamboo1,
				"38八万", "35五万", "27六索", "44红中",
				"22一索", "34四万", "35五万", "20八筒",
				"37七万", "13一筒", "43北风", "26五索",
				"21九筒", "25四索", "42西风",
			},
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
			Hands:         [4]Hand{{Concealed: []Tile{TileWindsWest}}},
		}
		err := round.Draw(DirectionEast)
		assert.NoError(t, err)
		assert.Equal(t, []Tile{"38八万", "35五万", "27六索", "44红中",
			"22一索", "34四万", "35五万", "20八筒",
			"37七万", "13一筒", "43北风", "26五索",
			"21九筒", "25四索", "42西风"}, round.Wall)
		assert.Equal(t, []Tile{TileWindsWest, TileBamboo1}, round.Hands[DirectionEast].Concealed)
		assert.Equal(t, DirectionEast, round.CurrentTurn)
		assert.Equal(t, ActionDiscard, round.CurrentAction)
	})
	t.Run("drawing flower", func(t *testing.T) {
		round := &Round{
			Wall: []Tile{TileGentlemen1, "38八万", "35五万", "27六索", "44红中",
				"22一索", "34四万", "35五万", "20八筒",
				"37七万", "13一筒", "43北风", "26五索",
				"21九筒", "25四索", "42西风", TileBamboo2},
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
			Hands:         [4]Hand{{Concealed: []Tile{TileWindsWest}}},
		}
		err := round.Draw(DirectionEast)
		assert.NoError(t, err)
		assert.Equal(t, []Tile{"38八万", "35五万", "27六索", "44红中",
			"22一索", "34四万", "35五万", "20八筒",
			"37七万", "13一筒", "43北风", "26五索",
			"21九筒", "25四索", "42西风"}, round.Wall)
		assert.Equal(t, []Tile{TileWindsWest, TileBamboo2}, round.Hands[DirectionEast].Concealed)
		assert.Equal(t, []Tile{TileGentlemen1}, round.Hands[DirectionEast].Flowers)
		assert.Equal(t, DirectionEast, round.CurrentTurn)
		assert.Equal(t, ActionDiscard, round.CurrentAction)
	})
	t.Run("drawing flower again", func(t *testing.T) {
		round := &Round{
			Wall: []Tile{TileGentlemen1, "38八万", "35五万", "27六索", "44红中",
				"22一索", "34四万", "35五万", "20八筒",
				"37七万", "13一筒", "43北风", "26五索",
				"21九筒", "25四索", "42西风", TileBamboo2, TileGentlemen2},
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
			Hands:         [4]Hand{{Concealed: []Tile{TileWindsWest}}},
		}
		err := round.Draw(DirectionEast)
		assert.NoError(t, err)
		assert.Equal(t, []Tile{"38八万", "35五万", "27六索", "44红中",
			"22一索", "34四万", "35五万", "20八筒",
			"37七万", "13一筒", "43北风", "26五索",
			"21九筒", "25四索", "42西风"}, round.Wall)
		assert.Equal(t, []Tile{TileWindsWest, TileBamboo2}, round.Hands[DirectionEast].Concealed)
		assert.Equal(t, []Tile{TileGentlemen1, TileGentlemen2}, round.Hands[DirectionEast].Flowers)
		assert.Equal(t, DirectionEast, round.CurrentTurn)
		assert.Equal(t, ActionDiscard, round.CurrentAction)
	})
}

func TestRound_Kong(t *testing.T) {
	t.Run("tiles exposed, tile from discards - cannot kong", func(t *testing.T) {
		round := &Round{
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
			Discards:      []Tile{TileBamboo1},
			Hands:         [4]Hand{{Revealed: [][]Tile{{TileBamboo1, TileBamboo1, TileBamboo1}}}},
		}
		err := round.Kong(DirectionEast, TileBamboo1)
		assert.EqualError(t, err, "not allowed")
	})
	t.Run("tiles concealed, tile from discards - exposed kong", func(t *testing.T) {
		round := &Round{
			Wall:          []Tile{TileWindsWest, TileWindsEast},
			CurrentTurn:   DirectionNorth,
			CurrentAction: ActionDraw,
			Discards:      []Tile{TileDragonsRed, TileBamboo1},
			Hands:         [4]Hand{{Concealed: []Tile{TileCharacters1, TileBamboo1, TileBamboo1, TileBamboo1, TileWindsWest}}},
		}
		err := round.Kong(DirectionEast, TileBamboo1)
		assert.NoError(t, err)
		assert.Equal(t, []Tile{TileDragonsRed}, round.Discards)
		assert.Equal(t, []Tile{TileWindsWest}, round.Wall)
		assert.Equal(t, []Tile{TileCharacters1, TileWindsWest, TileWindsEast}, round.Hands[DirectionEast].Concealed)
		assert.Equal(t, [][]Tile{{TileBamboo1, TileBamboo1, TileBamboo1, TileBamboo1}}, round.Hands[DirectionEast].Revealed)
		assert.Equal(t, DirectionEast, round.CurrentTurn)
		assert.Equal(t, ActionDiscard, round.CurrentAction)
	})
	t.Run("tiles exposed, tile from hand - exposed kong", func(t *testing.T) {
		round := &Round{
			Wall:          []Tile{TileWindsWest, TileWindsEast},
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDiscard,
			Discards:      []Tile{TileDragonsRed},
			Hands: [4]Hand{
				{
					Revealed:  [][]Tile{{TileBamboo1, TileBamboo1, TileBamboo1}},
					Concealed: []Tile{TileCharacters1, TileBamboo1, TileWindsWest}},
			},
		}
		err := round.Kong(DirectionEast, TileBamboo1)
		assert.NoError(t, err)
		assert.Equal(t, []Tile{TileDragonsRed}, round.Discards)
		assert.Equal(t, []Tile{TileWindsWest}, round.Wall)
		assert.Equal(t, []Tile{TileCharacters1, TileWindsWest, TileWindsEast}, round.Hands[DirectionEast].Concealed)
		assert.Equal(t, [][]Tile{{TileBamboo1, TileBamboo1, TileBamboo1, TileBamboo1}}, round.Hands[DirectionEast].Revealed)
		assert.Equal(t, DirectionEast, round.CurrentTurn)
		assert.Equal(t, ActionDiscard, round.CurrentAction)
	})
	t.Run("tiles concealed, tile from hand - hidden kong", func(t *testing.T) {
		round := &Round{
			Wall:          []Tile{TileWindsWest, TileWindsEast},
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDiscard,
			Discards:      []Tile{TileDragonsRed},
			Hands: [4]Hand{
				{
					Concealed: []Tile{TileCharacters1, TileBamboo1, TileBamboo1, TileBamboo1, TileBamboo1, TileWindsWest}},
			},
		}
		err := round.Kong(DirectionEast, TileBamboo1)
		assert.NoError(t, err)
		assert.Equal(t, []Tile{TileDragonsRed}, round.Discards)
		assert.Equal(t, []Tile{TileWindsWest}, round.Wall)
		assert.Equal(t, []Tile{TileCharacters1, TileWindsWest, TileWindsEast}, round.Hands[DirectionEast].Concealed)
		assert.Equal(t, [][]Tile{{TileBamboo1, TileBamboo1, TileBamboo1, TileBamboo1}}, round.Hands[DirectionEast].Revealed)
		assert.Equal(t, DirectionEast, round.CurrentTurn)
		assert.Equal(t, ActionDiscard, round.CurrentAction)
	})
	t.Run("drawing flower after kong", func(t *testing.T) {
		round := &Round{
			Wall:          []Tile{TileWindsWest, TileWindsEast, TileCat},
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDiscard,
			Discards:      []Tile{TileDragonsRed},
			Hands: [4]Hand{
				{
					Concealed: []Tile{TileCharacters1, TileBamboo1, TileBamboo1, TileBamboo1, TileBamboo1, TileWindsWest}},
			},
		}
		err := round.Kong(DirectionEast, TileBamboo1)
		assert.NoError(t, err)
		assert.Equal(t, []Tile{TileDragonsRed}, round.Discards)
		assert.Equal(t, []Tile{TileWindsWest}, round.Wall)
		assert.Equal(t, []Tile{TileCat}, round.Hands[DirectionEast].Flowers)
		assert.Equal(t, []Tile{TileCharacters1, TileWindsWest, TileWindsEast}, round.Hands[DirectionEast].Concealed)
	})
}

func TestRound_Win(t *testing.T) {
	t.Run("can't win from discards when there are no discards", func(t *testing.T) {
		round := &Round{
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
		}
		err := round.Win(DirectionEast, [][]Tile{{TileCharacters4, TileCharacters4}})
		assert.Error(t, err)
	})
	t.Run("dan diao, win from discards", func(t *testing.T) {
		round := &Round{
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
			Discards:      []Tile{TileCharacters3, TileCharacters4},
			Hands: [4]Hand{
				{
					Revealed: [][]Tile{
						{TileBamboo3, TileBamboo4, TileBamboo5},
						{TileDots5, TileDots6, TileDots7},
						{TileWindsNorth, TileWindsNorth, TileWindsNorth},
						{TileDragonsWhite, TileDragonsWhite, TileDragonsWhite},
					},
					Concealed: []Tile{TileCharacters4},
				},
			},
		}
		err := round.Win(DirectionEast, [][]Tile{{TileCharacters4, TileCharacters4}})
		assert.NoError(t, err)
		assert.Equal(t, []Tile{TileCharacters3}, round.Discards)
		assert.Equal(t, ActionGameOver, round.CurrentAction)
		assert.Equal(t, DirectionEast, round.CurrentTurn)
		assert.Equal(t, [][]Tile{
			{TileBamboo3, TileBamboo4, TileBamboo5},
			{TileDots5, TileDots6, TileDots7},
			{TileWindsNorth, TileWindsNorth, TileWindsNorth},
			{TileDragonsWhite, TileDragonsWhite, TileDragonsWhite},
			{TileCharacters4, TileCharacters4},
		}, round.Hands[DirectionEast].Revealed)
		assert.Empty(t, round.Hands[DirectionEast].Concealed)
	})
	t.Run("invalid melds", func(t *testing.T) {
		round := &Round{
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
			Discards:      []Tile{TileCharacters4},
			Hands: [4]Hand{
				{
					Revealed: [][]Tile{
						{TileBamboo3, TileBamboo4, TileBamboo5},
						{TileDots5, TileDots6, TileDots7},
						{TileWindsNorth, TileWindsNorth, TileWindsNorth},
					},
					Concealed: []Tile{TileCharacters4, TileDragonsRed, TileDragonsWhite, TileDragonsWhite},
				},
			},
		}
		err := round.Win(DirectionEast, [][]Tile{{TileCharacters4, TileCharacters4}, {TileDragonsRed, TileDragonsWhite, TileDragonsWhite}})
		assert.Error(t, err)
	})
	t.Run("not enough tiles", func(t *testing.T) {
		round := &Round{
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
			Discards:      []Tile{TileCharacters4},
			Hands: [4]Hand{
				{
					Revealed: [][]Tile{
						{TileBamboo3, TileBamboo4, TileBamboo5},
						{TileDots5, TileDots6, TileDots7},
						{TileWindsNorth, TileWindsNorth, TileWindsNorth},
					},
					Concealed: []Tile{TileCharacters4, TileDragonsWhite, TileDragonsWhite},
				},
			},
		}
		err := round.Win(DirectionEast, [][]Tile{{TileCharacters4, TileCharacters4}, {TileDragonsWhite, TileDragonsWhite, TileDragonsWhite}})
		assert.Error(t, err)

	})
	t.Run("not enough melds", func(t *testing.T) {
		round := &Round{
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDraw,
			Discards:      []Tile{TileCharacters4},
			Hands: [4]Hand{
				{
					Revealed: [][]Tile{
						{TileBamboo3, TileBamboo4, TileBamboo5},
						{TileDots5, TileDots6, TileDots7},
						{TileWindsNorth, TileWindsNorth, TileWindsNorth},
					},
					Concealed: []Tile{TileCharacters4, TileDragonsRed, TileDragonsWhite, TileDragonsWhite, TileDragonsWhite},
				},
			},
		}
		err := round.Win(DirectionEast, [][]Tile{{TileCharacters4, TileCharacters4}})
		assert.Error(t, err)
	})
	t.Run("dan diao zi mo", func(t *testing.T) {
		round := &Round{
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDiscard,
			Hands: [4]Hand{
				{
					Revealed: [][]Tile{
						{TileBamboo3, TileBamboo4, TileBamboo5},
						{TileDots5, TileDots6, TileDots7},
						{TileWindsNorth, TileWindsNorth, TileWindsNorth},
						{TileDragonsWhite, TileDragonsWhite, TileDragonsWhite},
					},
					Concealed: []Tile{TileCharacters4, TileCharacters4},
				},
			},
		}
		err := round.Win(DirectionEast, [][]Tile{{TileCharacters4, TileCharacters4}})
		assert.NoError(t, err)
		assert.Equal(t, ActionGameOver, round.CurrentAction)
		assert.Equal(t, DirectionEast, round.CurrentTurn)
		assert.Equal(t, [][]Tile{
			{TileBamboo3, TileBamboo4, TileBamboo5},
			{TileDots5, TileDots6, TileDots7},
			{TileWindsNorth, TileWindsNorth, TileWindsNorth},
			{TileDragonsWhite, TileDragonsWhite, TileDragonsWhite},
			{TileCharacters4, TileCharacters4},
		}, round.Hands[DirectionEast].Revealed)
		assert.Empty(t, round.Hands[DirectionEast].Concealed)
	})
}

func TestRound_EndGame(t *testing.T) {
	t.Run("game is a draw", func(t *testing.T) {
		round := &Round{
			Wall: []Tile{
				"38八万", "35五万", "27六索", "44红中",
				"22一索", "34四万", "35五万", "20八筒",
				"37七万", "13一筒", "43北风", "26五索",
				"21九筒", "25四索", "42西风",
			},
			CurrentTurn:   DirectionEast,
			CurrentAction: ActionDiscard,
		}
		err := round.EndGame(DirectionEast)
		assert.NoError(t, err)
	})
}
