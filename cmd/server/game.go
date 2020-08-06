package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/yi-jiayu/mahjong.go"
)

type Event struct {
	SequenceNumber int      `json:"sequence_number"`
	Seat           int      `json:"seat"`
	Action         string   `json:"action"`
	Tiles          []string `json:"tiles"`
}

type Game struct {
	Round *mahjong.Round

	// Client connections registry
	clients map[chan string]struct{}

	sync.Mutex
}

func (g *Game) handleEvent(e Event) {
	g.Lock()
	defer g.Unlock()

	if e.SequenceNumber != g.Round.SequenceNumber {
		// Ignore events with wrong sequence number
		return
	}

	switch e.Action {
	case mahjong.ActionDiscard:
		if len(e.Tiles) < 1 {
			return
		}
		g.Round.Discard(e.Seat, e.Tiles[0])
	case "chow":
		if len(e.Tiles) < 2 {
			return
		}
		g.Round.Chow(e.Seat, e.Tiles[0], e.Tiles[1])
	case "peng":
		if len(e.Tiles) < 1 {
			return
		}
		g.Round.Peng(e.Seat, e.Tiles[0])
	case "draw":
		g.Round.Draw(e.Seat)
	case "reset":
		g.Round = mahjong.NewRound(0, mahjong.DirectionEast)
	}

	// We got a new event from the outside!
	// Send event to all connected clients
	var b bytes.Buffer
	json.NewEncoder(&b).Encode(g.Round)
	for clientMessageChan := range g.clients {
		clientMessageChan <- b.String()
	}
}

func (g *Game) addClient(c chan string) {
	g.Lock()
	defer g.Unlock()
	g.clients[c] = struct{}{}
}

func (g *Game) removeClient(c chan string) {
	g.Lock()
	defer g.Unlock()
	delete(g.clients, c)
}

func (g *Game) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	c := make(chan string)

	// Signal the broker that we have a new connection
	g.addClient(c)

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer g.removeClient(c)

	// Listen to connection close and un-register c
	notify := r.Context().Done()
	go func() {
		<-notify
		g.removeClient(c)
	}()

	for {
		round := <-c
		fmt.Fprintf(w, "data: %s\n\n", round)
		flusher.Flush()
	}
}

func sendMockEvents(game *Game) {
	events := []Event{
		{
			Seat:   mahjong.DirectionEast,
			Action: "discard",
			Tiles:  []string{mahjong.TileBamboo1},
		},
		{
			Seat:   mahjong.DirectionSouth,
			Action: "peng",
			Tiles:  []string{mahjong.TileBamboo1},
		},
		{
			Seat:   mahjong.DirectionSouth,
			Action: "discard",
			Tiles:  []string{mahjong.TileCharacters2},
		},
		{
			Seat:   mahjong.DirectionWest,
			Action: "draw",
		},
		{
			Seat:   mahjong.DirectionWest,
			Action: "discard",
			Tiles:  []string{mahjong.TileCharacters7},
		},
		{
			Seat:   mahjong.DirectionSouth,
			Action: "peng",
			Tiles:  []string{mahjong.TileCharacters7},
		},
		{
			Seat:   mahjong.DirectionSouth,
			Action: "discard",
			Tiles:  []string{mahjong.TileBamboo4},
		},
		{
			Seat:   mahjong.DirectionWest,
			Action: "draw",
		},
		{
			Seat:   mahjong.DirectionWest,
			Action: "discard",
			Tiles:  []string{mahjong.TileBamboo9},
		},
		{
			Seat:   mahjong.DirectionNorth,
			Action: "draw",
		},
		{
			Seat:   mahjong.DirectionNorth,
			Action: "discard",
			Tiles:  []string{mahjong.TileDots9},
		},
		{
			Action: "reset",
		},
	}

	for {
		i := 0
		for _, e := range events {
			time.Sleep(time.Second * 2)
			e.SequenceNumber = i
			game.handleEvent(e)
			i++
		}
	}
}

func main() {
	game := &Game{
		Round:   mahjong.NewRound(0, mahjong.DirectionEast),
		clients: make(map[chan string]struct{}),
	}

	go sendMockEvents(game)

	log.Fatal("HTTP server error: ", http.ListenAndServe("localhost:3000", game))
}
