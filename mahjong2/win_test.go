package mahjong2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_search(t *testing.T) {
	t.Run("hand with multiple winning combinations", func(t *testing.T) {
		tiles := NewTileBag([]Tile{
			TileDots1, TileDots1, TileDots1,
			TileDots2, TileDots2, TileDots2,
			TileDots3, TileDots3, TileDots3,
			TileDragonsWhite, TileDragonsWhite,
		})
		result := search(tiles)
		assert.Equal(t, [][]Meld{
			{
				{Type: MeldChi, Tiles: []Tile{"13一筒", "14二筒", "15三筒"}},
				{Type: MeldChi, Tiles: []Tile{"13一筒", "14二筒", "15三筒"}},
				{Type: MeldChi, Tiles: []Tile{"13一筒", "14二筒", "15三筒"}},
				{Type: MeldEyes, Tiles: []Tile{"46白板"}},
			},
			{
				{Type: MeldPong, Tiles: []Tile{"13一筒"}},
				{Type: MeldPong, Tiles: []Tile{"14二筒"}},
				{Type: MeldPong, Tiles: []Tile{"15三筒"}},
				{Type: MeldEyes, Tiles: []Tile{"46白板"}},
			},
		}, result)
	})
	t.Run("hand with odd number of tiles", func(t *testing.T) {
		tiles := NewTileBag([]Tile{
			TileDots1, TileDots1, TileDots1,
			TileDots2, TileDots2, TileDots2,
			TileDots3, TileDots3, TileDots3,
			TileDragonsWhite,
		})
		result := search(tiles)
		assert.Empty(t, result)
	})
	t.Run("eyes only", func(t *testing.T) {
		tiles := NewTileBag([]Tile{
			TileDragonsWhite, TileDragonsWhite,
		})
		result := search(tiles)
		assert.Equal(t, [][]Meld{{{Type: MeldEyes, Tiles: []Tile{"46白板"}}}}, result)
	})
}

func Benchmark_search(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tiles := NewTileBag([]Tile{
			TileDots1, TileDots1, TileDots1,
			TileDots2, TileDots2, TileDots2,
			TileDots3, TileDots3, TileDots3,
			TileDragonsWhite, TileDragonsWhite,
		})
		search(tiles)
	}
}

func Test_score(t *testing.T) {
	t.Run("ping hu", func(t *testing.T) {
		melds := []Meld{
			{Type: MeldChi, Tiles: []Tile{TileDots1, TileDots2, TileDots3}},
			{Type: MeldChi, Tiles: []Tile{TileCharacters4, TileCharacters5, TileCharacters6}},
			{Type: MeldChi, Tiles: []Tile{TileBamboo2, TileBamboo3, TileBamboo4}},
			{Type: MeldEyes, Tiles: []Tile{TileCharacters9, TileCharacters9}},
		}
		score(nil, 0, melds)
	})
}
