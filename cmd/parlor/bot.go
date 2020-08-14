package main

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/yi-jiayu/mahjong.go"
)

type Bot struct {
	ID          string
	Room        *Room
	GameUpdates chan string
}

func NewBot(room *Room) *Bot {
	p := Player{
		NamePrefix: "bot",
	}
	_ = playerRepository.Insert(&p)
	return &Bot{
		ID:          p.ID,
		Room:        room,
		GameUpdates: make(chan string),
	}
}

func (b *Bot) Run() {
	for update := range b.GameUpdates {
		go func(update string) {
			var state struct {
				Seat    mahjong.Direction  `json:"seat"`
				Nonce   int                `json:"nonce"`
				Phase   int                `json:"phase"`
				Players []string           `json:"players"`
				Round   *mahjong.RoundView `json:"round"`
			}
			json.Unmarshal([]byte(update), &state)
			if state.Phase != PhaseInProgress {
				return
			}
			if state.Round.CurrentTurn == state.Seat {
				if state.Round.CurrentAction == mahjong.ActionDraw {
					delay := 2 + rand.Intn(4)
					time.Sleep(time.Duration(delay) * time.Second)
					b.Room.HandleAction(b.ID, Action{
						Nonce: state.Nonce,
						Type:  "draw",
					})
				} else if state.Round.CurrentAction == mahjong.ActionDiscard {
					concealed := state.Round.Hands[state.Seat].Concealed
					tileToDiscard := concealed[rand.Intn(len(concealed))]
					b.Room.HandleAction(b.ID, Action{
						Nonce: state.Nonce,
						Type:  "discard",
						Tiles: []mahjong.Tile{tileToDiscard},
					})
				}
			}
		}(update)
	}
}
