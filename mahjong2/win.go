package mahjong2

import (
	"fmt"
	"sort"
)

type searchState struct {
	tiles TileBag
	melds []Meld
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
	sort.Slice(s.melds, func(i, j int) bool {
		if s.melds[i].Type < s.melds[j].Type {
			return true
		}
		return s.melds[i].Tiles[0] < s.melds[j].Tiles[0]
	})
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
