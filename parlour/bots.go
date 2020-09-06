package parlour

import (
	"errors"
	"fmt"
	"time"

	"github.com/yi-jiayu/mahjong.go"
)

type Bot struct {
	ID      string
	Room    *Room
	Updates chan RoomView
}

func (b *Bot) Start(roomService *roomService) {
	for view := range b.Updates {
		go func(view RoomView) {
			if view.Round == nil {
				return
			}
			if view.Round.Turn != view.Round.Seat {
				return
			}
			if view.Round.Phase == mahjong.PhaseDraw {
				time.Sleep(time.Duration(view.Round.ReservedDuration)*time.Millisecond + time.Second)
				action := Action{
					Nonce: view.Nonce,
					Type:  ActionDraw,
				}
				err := roomService.Dispatch(b.Room, b.ID, action)
				if err != nil && !errors.Is(err, errInvalidNonce) {
					fmt.Printf("error drawing: %v", err)
					return
				}
			} else if view.Round.Phase == mahjong.PhaseDiscard {
				if view.Round.DrawsLeft <= 0 && !view.Round.Finished {
					action := Action{
						Nonce: view.Nonce,
						Type:  ActionEndRound,
					}
					err := roomService.Dispatch(b.Room, b.ID, action)
					if err != nil && !errors.Is(err, errInvalidNonce) {
						fmt.Printf("error ending round: %v", err)
						return
					}
					return
				}
				var tileToDiscard mahjong.Tile
				for tile := range view.Round.Hands[view.Round.Seat].Concealed {
					tileToDiscard = tile
					break
				}
				action := Action{
					Nonce: view.Nonce,
					Type:  ActionDiscard,
					Tiles: []mahjong.Tile{tileToDiscard},
				}
				err := roomService.Dispatch(b.Room, b.ID, action)
				if err != nil && !errors.Is(err, errInvalidNonce) {
					fmt.Printf("error discarding: %v", err)
					return
				}
			}
		}(view)
	}
}
