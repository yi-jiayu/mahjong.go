package parlour

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/yi-jiayu/mahjong.go/mahjong"
)

type Bot struct {
	ID      string
	Room    *Room
	Updates chan string
}

func (b *Bot) Start() {
	for update := range b.Updates {
		go func(update string) {
			var state RoomView
			err := json.Unmarshal([]byte(update), &state)
			if err != nil {
				fmt.Printf("error unmarshalling game state: %v\n", err)
				return
			}
			if state.Round.Turn != state.Round.Seat {
				return
			}
			if state.Round.Phase == mahjong.PhaseDraw {
				time.Sleep(time.Duration(state.Round.ReservedDuration)*time.Millisecond + time.Second)
				b.Room.WithLock(func(r *Room) {
					err := r.reduce(b.ID, Action{
						Nonce: state.Nonce,
						Type:  ActionDraw,
					})
					if err != nil && !errors.Is(err, errInvalidNonce) {
						fmt.Printf("error drawing: %v", err)
						return
					}
				})
			} else if state.Round.Phase == mahjong.PhaseDiscard {
				if state.Round.DrawsLeft <= 0 {
					b.Room.WithLock(func(r *Room) {
						err := r.reduce(b.ID, Action{
							Nonce: state.Nonce,
							Type:  ActionEndRound,
						})
						if err != nil && !errors.Is(err, errInvalidNonce) {
							fmt.Printf("error ending round: %v", err)
							return
						}
					})
					return
				}
				var tileToDiscard mahjong.Tile
				for tile := range state.Round.Hands[state.Round.Seat].Concealed {
					tileToDiscard = tile
					break
				}
				b.Room.WithLock(func(r *Room) {
					err := r.reduce(b.ID, Action{
						Nonce: state.Nonce,
						Type:  ActionDiscard,
						Tiles: []mahjong.Tile{tileToDiscard},
					})
					if err != nil && !errors.Is(err, errInvalidNonce) {
						fmt.Printf("error discarding: %v", err)
						return
					}
				})
			}
		}(update)
	}
}
