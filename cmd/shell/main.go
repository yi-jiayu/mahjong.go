package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/yi-jiayu/mahjong.go"
)

func prompt(round *mahjong.Round) {
	fmt.Printf("Prevailing wind: %d, Turn: %d, Action: %s\n", round.PrevailingWind, round.CurrentTurn, round.CurrentAction)
	hand := round.Hands[round.CurrentTurn]
	concealed := hand.Concealed
	sort.Strings(concealed)
	fmt.Printf("Discards: %v\nFlowers: %v\nRevealed: %v\nConcealed: %v\n", round.Discards, hand.Flowers, hand.Revealed, concealed)
	fmt.Print("> ")
}

func main() {
	round := mahjong.NewRound(time.Now().UnixNano(), mahjong.DirectionEast, mahjong.DirectionEast)
	s := bufio.NewScanner(os.Stdin)
	prompt(round)
	for s.Scan() {
		text := s.Text()

		switch {
		case strings.HasPrefix(text, "discard "):
			tile := strings.TrimPrefix(text, "discard ")
			err := round.Discard(round.CurrentTurn, tile)
			if err != nil {
				fmt.Println(err)
			}
		case text == "draw":
			err := round.Draw(round.CurrentTurn)
			if err != nil {
				fmt.Println(err)
			}
		case strings.HasPrefix(text, "chow "):
			tiles := strings.Fields(strings.TrimPrefix(text, "chow "))
			if len(tiles) < 2 {
				fmt.Println("not enough arguments")
				continue
			}
			err := round.Chow(round.CurrentTurn, tiles[0], tiles[1])
			if err != nil {
				fmt.Println(err)
			}
		case strings.HasPrefix(text, "peng "):
			tile := strings.TrimPrefix(text, "peng ")
			err := round.Peng(round.CurrentTurn, tile)
			if err != nil {
				fmt.Println(err)
			}
		}

		fmt.Println()
		prompt(round)
	}
}
