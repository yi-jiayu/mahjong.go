package mahjong2

import (
	"fmt"
	"sort"
)

type searchState struct {
	tiles TileBag
	melds Melds
}

func (s searchState) copy() searchState {
	var cpy searchState
	cpy.tiles = TileBag{}
	for tile, count := range s.tiles {
		cpy.tiles[tile] = count
	}
	cpy.melds = make([]Meld, len(s.melds))
	for i, melds := range s.melds {
		cpy.melds[i].Type = melds.Type
		cpy.melds[i].Tiles = make([]Tile, len(melds.Tiles))
		copy(cpy.melds[i].Tiles, melds.Tiles)
	}
	return cpy
}

func (s searchState) hash() string {
	sort.Sort(s.melds)
	return fmt.Sprint(s)
}

func pop(stack []searchState) (searchState, []searchState) {
	return stack[len(stack)-1], stack[:len(stack)-1]
}

func push(stack []searchState, state searchState) []searchState {
	return append(stack, state)
}

func search(tiles TileBag) [][]Meld {
	var results [][]Meld
	seen := make(map[string]struct{})
	stack := []searchState{{tiles: tiles}}
	for len(stack) > 0 {
		var state searchState
		state, stack = pop(stack)
		hash := state.hash()
		if _, ok := seen[hash]; ok {
			continue
		}
		seen[hash] = struct{}{}
		// check for eyes
		if len(state.tiles) == 1 {
			for tile, count := range state.tiles {
				if count == 2 {
					melds := append(state.melds, Meld{
						Type:  MeldEyes,
						Tiles: []Tile{tile},
					})
					results = append(results, melds)
					continue
				}
			}
		}
		for tile := range state.tiles {
			// check for pongs
			if state.tiles.Count(tile) > 2 {
				s := state.copy()
				s.tiles.RemoveN(tile, 3)
				s.melds = append(s.melds, Meld{
					Type:  MeldPong,
					Tiles: []Tile{tile},
				})
				stack = push(stack, s)
			}
			// check for chi
			if connecting, ok := sequences[tile]; ok {
				for _, c := range connecting {
					if state.tiles.Contains(c[0]) && state.tiles.Contains(c[1]) {
						s := state.copy()
						seq := []Tile{tile, c[0], c[1]}
						sort.Slice(seq, func(i, j int) bool {
							return seq[i] < seq[j]
						})
						s.tiles.Remove(tile)
						s.tiles.Remove(c[0])
						s.tiles.Remove(c[1])
						s.melds = append(s.melds, Meld{
							Type:  MeldChi,
							Tiles: seq,
						})
						stack = push(stack, s)
					}
				}
			}
		}
	}
	return results
}

func isFlowerForSeat(flower Tile, seat int) bool {
	if flower == TileCat || flower == TileRat || flower == TileRooster || flower == TileCentipede {
		return true
	}
	switch seat {
	case 0:
		if flower == TileGentlemen1 || flower == TileSeasons1 {
			return true
		}
	case 1:
		if flower == TileGentlemen2 || flower == TileSeasons2 {
			return true
		}
	case 2:
		if flower == TileGentlemen3 || flower == TileSeasons3 {
			return true
		}
	case 3:
		if flower == TileGentlemen4 || flower == TileSeasons4 {
			return true
		}
	}
	return false
}

func isMatchingWind(wind Tile, seat Direction) bool {
	switch {
	case wind == TileWindsEast && seat == DirectionEast:
		return true
	case wind == TileWindsSouth && seat == DirectionWest:
		return true
	case wind == TileWindsWest && seat == DirectionSouth:
		return true
	case wind == TileWindsNorth && seat == DirectionNorth:
		return true
	}
	return false
}

func isFullFlush(suits map[Suit]int) bool {
	return len(suits) == 1
}

func isHalfFlush(suits map[Suit]int) bool {
	cardinality := len(suits)
	if suits[SuitDragons] > 0 {
		cardinality--
	}
	if suits[SuitWinds] > 0 {
		cardinality--
	}
	return cardinality == 1
}

func score(round *Round, seat int, melds Melds) int {
	score := 0
	// zi mo
	if round.Turn == seat {
		score++
	}
	meldTypes := make(map[MeldType]int)
	suits := make(map[Suit]int)
	for _, meld := range melds {
		meldTypes[meld.Type]++
		suits[meld.Tiles[0].Suit()]++
	}
	if isFullFlush(suits) {
		score += 4
	} else if isHalfFlush(suits) {
		score += 2
	}
	// ping hu
	if meldTypes[MeldChi] == 4 {
		// no flowers
		if len(round.Hands[seat].Flowers) == 0 {
			return score + 4
		}
		// chou ping hu
		return score + 1
	}
	// pong pong hu
	if meldTypes[MeldPong]+meldTypes[MeldGang] == 4 {
		score += 2
	}
	// flowers
	for _, flower := range round.Hands[seat].Flowers {
		if isFlowerForSeat(flower, seat) {
			score++
		}
	}
	for _, m := range melds {
		if m.Type == MeldPong || m.Type == MeldGang {
			if m.Tiles[0] == TileDragonsRed || m.Tiles[0] == TileDragonsGreen || m.Tiles[0] == TileDragonsWhite {
				score++
			}
			if isMatchingWind(m.Tiles[0], round.seatWind(seat)) {
				score++
			}
			if isMatchingWind(m.Tiles[0], round.Wind) {
				score++
			}
		}
	}
	return score
}
