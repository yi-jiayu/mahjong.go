package mahjong2

// Tile represents a mahjong tile.
type Tile string

// All tiles.
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

var flowers = []Tile{"01猫", "02老鼠", "03公鸡", "04蜈蚣", "05梅", "06兰", "07菊", "08竹", "09春", "10夏", "11秋", "12冬"}

func isFlower(tile Tile) bool {
	for _, flower := range flowers {
		if tile == flower {
			return true
		}
	}
	return false
}

// sequences is a map of tiles to valid tiles for completing a sequence.
var sequences = map[Tile][][2]Tile{
	TileDots1:       {{TileDots2, TileDots3}},
	TileDots2:       {{TileDots1, TileDots3}, {TileDots3, TileDots4}},
	TileDots3:       {{TileDots1, TileDots2}, {TileDots2, TileDots4}, {TileDots4, TileDots5}},
	TileDots4:       {{TileDots2, TileDots4}, {TileDots4, TileDots5}, {TileDots5, TileDots6}},
	TileDots5:       {{TileDots4, TileDots5}, {TileDots5, TileDots6}, {TileDots6, TileDots7}},
	TileDots6:       {{TileDots5, TileDots6}, {TileDots6, TileDots7}, {TileDots7, TileDots8}},
	TileDots7:       {{TileDots6, TileDots7}, {TileDots7, TileDots8}, {TileDots8, TileDots9}},
	TileDots8:       {{TileDots7, TileDots8}, {TileDots8, TileDots9}},
	TileDots9:       {{TileDots8, TileDots9}},
	TileBamboo1:     {{TileBamboo2, TileBamboo3}},
	TileBamboo2:     {{TileBamboo1, TileBamboo3}, {TileBamboo3, TileBamboo4}},
	TileBamboo3:     {{TileBamboo1, TileBamboo2}, {TileBamboo2, TileBamboo4}, {TileBamboo4, TileBamboo5}},
	TileBamboo4:     {{TileBamboo2, TileBamboo4}, {TileBamboo4, TileBamboo5}, {TileBamboo5, TileBamboo6}},
	TileBamboo5:     {{TileBamboo4, TileBamboo5}, {TileBamboo5, TileBamboo6}, {TileBamboo6, TileBamboo7}},
	TileBamboo6:     {{TileBamboo5, TileBamboo6}, {TileBamboo6, TileBamboo7}, {TileBamboo7, TileBamboo8}},
	TileBamboo7:     {{TileBamboo6, TileBamboo7}, {TileBamboo7, TileBamboo8}, {TileBamboo8, TileBamboo9}},
	TileBamboo8:     {{TileBamboo7, TileBamboo8}, {TileBamboo8, TileBamboo9}},
	TileBamboo9:     {{TileBamboo8, TileBamboo9}},
	TileCharacters1: {{TileCharacters2, TileCharacters3}},
	TileCharacters2: {{TileCharacters1, TileCharacters3}, {TileCharacters3, TileCharacters4}},
	TileCharacters3: {{TileCharacters1, TileCharacters2}, {TileCharacters2, TileCharacters4}, {TileCharacters4, TileCharacters5}},
	TileCharacters4: {{TileCharacters2, TileCharacters4}, {TileCharacters4, TileCharacters5}, {TileCharacters5, TileCharacters6}},
	TileCharacters5: {{TileCharacters4, TileCharacters5}, {TileCharacters5, TileCharacters6}, {TileCharacters6, TileCharacters7}},
	TileCharacters6: {{TileCharacters5, TileCharacters6}, {TileCharacters6, TileCharacters7}, {TileCharacters7, TileCharacters8}},
	TileCharacters7: {{TileCharacters6, TileCharacters7}, {TileCharacters7, TileCharacters8}, {TileCharacters8, TileCharacters9}},
	TileCharacters8: {{TileCharacters7, TileCharacters8}, {TileCharacters8, TileCharacters9}},
	TileCharacters9: {{TileCharacters8, TileCharacters9}},
}
