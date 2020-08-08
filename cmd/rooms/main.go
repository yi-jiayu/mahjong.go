package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"

	"github.com/yi-jiayu/mahjong.go"
)

type PlayerRegistry struct {
	sync.RWMutex
	Names map[string]string
}

func (r *PlayerRegistry) GetName(id string) string {
	r.RLock()
	defer r.RUnlock()
	return r.Names[id]
}

type RoomRegistry struct {
	sync.RWMutex
	Rooms map[string]*Room
}

func (r *RoomRegistry) GetRoom(id string) *Room {
	r.RLock()
	defer r.RUnlock()
	return r.Rooms[id]
}

var (
	playerRegistry = &PlayerRegistry{
		Names: map[string]string{},
	}
	roomRegistry = &RoomRegistry{
		Rooms: map[string]*Room{},
	}
)

const (
	PhaseLobby = iota
	PhaseInProgress
)

type Room struct {
	Phase   int
	Players []string
	Round   *mahjong.Round

	sync.Mutex
	clients map[chan string]struct{}
}

func (r *Room) MarshalJSON() ([]byte, error) {
	players := make([]string, len(r.Players))
	for i, playerID := range r.Players {
		players[i] = playerRegistry.GetName(playerID)
	}
	return json.Marshal(struct {
		Phase   int            `json:"phase"`
		Players []string       `json:"players"`
		Round   *mahjong.Round `json:"round"`
	}{
		Phase:   r.Phase,
		Players: players,
		Round:   r.Round,
	})
}

func newSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func newRoom(players ...string) string {
	roomRegistry.Lock()
	defer roomRegistry.Unlock()

	charset := "ABCDEFGHIJKLMNOOPQRSTUVWXYZ"
	for {
		id := fmt.Sprintf("%c%c%c%c", charset[rand.Intn(len(charset))], charset[rand.Intn(len(charset))], charset[rand.Intn(len(charset))], charset[rand.Intn(len(charset))])
		if _, ok := roomRegistry.Rooms[id]; !ok {
			roomRegistry.Rooms[id] = &Room{
				Players: players,
				clients: map[chan string]struct{}{},
			}
			return id
		}
	}
}

func (r *Room) addClient(c chan string) {
	r.Lock()
	defer r.Unlock()

	r.clients[c] = struct{}{}
	var b bytes.Buffer
	json.NewEncoder(&b).Encode(r)
	c <- b.String()
}

func (r *Room) removeClient(c chan string) {
	r.Lock()
	defer r.Unlock()

	delete(r.clients, c)
}

func (r *Room) broadcast() {
	r.Lock()
	defer r.Unlock()

	var b bytes.Buffer
	json.NewEncoder(&b).Encode(r)
	for c := range r.clients {
		c <- b.String()
	}
}

func (r *Room) StartRound() error {
	r.Lock()
	defer r.Unlock()
	if len(r.Players) < 4 {
		return errors.New("not enough players")
	}
	r.Round = mahjong.NewRound(0, mahjong.DirectionEast)
	r.Round.Deal()
	r.Phase = PhaseInProgress
	return nil
}

func main() {
	r := gin.Default()
	store := memstore.NewStore([]byte("secret"))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,
	}))
	r.Use(sessions.Sessions("MJSESSIONID", store))
	r.Use(func(c *gin.Context) {
		session := sessions.Default(c)
		if id := session.Get("id"); id != nil {
			c.Set("id", id)
		} else {
			id := newSessionID()
			session.Set("id", id)
			playerRegistry.Lock()
			for {
				name := fmt.Sprintf("anon#%04d", rand.Intn(1000))
				if _, ok := playerRegistry.Names[id]; !ok {
					playerRegistry.Names[id] = name
					break
				}
			}
			playerRegistry.Unlock()
			session.Save()
			c.Set("id", id)
		}
		c.Next()
	})
	r.GET("/self", func(c *gin.Context) {
		id := c.MustGet("id").(string)
		name := playerRegistry.GetName(id)
		c.JSON(http.StatusOK, map[string]string{
			"id":   id,
			"name": name,
		})
	})
	r.POST("/rooms", func(c *gin.Context) {
		id := c.MustGet("id").(string)
		roomID := newRoom(id)
		c.JSON(http.StatusOK, map[string]string{
			"room_id": roomID,
		})
	})
	r.GET("/rooms/:id/live", func(c *gin.Context) {
		roomID := strings.ToUpper(c.Param("id"))
		room := roomRegistry.GetRoom(roomID)
		if room == nil {
			c.String(http.StatusNotFound, "Not Found")
			return
		}
		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.String(http.StatusInternalServerError, "Streaming Unsupported")
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		// Each connection registers its own message channel with the Broker's connections registry
		ch := make(chan string)

		// Signal the broker that we have a new connection
		go room.addClient(ch)

		// Remove this client from the map of connected clients
		// when this handler exits.
		defer room.removeClient(ch)

		// Listen to connection close and un-register c
		notify := c.Request.Context().Done()
		go func() {
			<-notify
			room.removeClient(ch)
		}()

		for {
			game := <-ch
			fmt.Fprintf(c.Writer, "data: %s\n\n", game)
			flusher.Flush()
		}
	})
	r.POST("/rooms/:id/players", func(c *gin.Context) {
		roomID := strings.ToUpper(c.Param("id"))
		playerID := c.MustGet("id").(string)
		room := roomRegistry.GetRoom(roomID)
		if room == nil {
			c.String(http.StatusNotFound, "Not Found")
			return
		}
		if len(room.Players) == 4 {
			c.String(http.StatusBadRequest, "Room Full")
			return
		}
		for _, p := range room.Players {
			if p == playerID {
				c.String(http.StatusBadRequest, "Already Joined")
				return
			}
		}
		room.Players = append(room.Players, playerID)
		room.broadcast()
	})
	r.POST("/rooms/:id/start", func(c *gin.Context) {
		roomID := strings.ToUpper(c.Param("id"))
		room := roomRegistry.GetRoom(roomID)
		if room == nil {
			c.String(http.StatusNotFound, "Not Found")
			return
		}
		playerID := c.MustGet("id").(string)
		found := false
		for _, id := range room.Players {
			if id == playerID {
				found = true
			}
		}
		if !found {
			c.String(http.StatusForbidden, "Not In Room")
			return
		}
		err := room.StartRound()
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		room.broadcast()
	})
	r.GET("/rooms/:id/self", func(c *gin.Context) {
		roomID := strings.ToUpper(c.Param("id"))
		room := roomRegistry.GetRoom(roomID)
		if room == nil {
			c.String(http.StatusNotFound, "Not Found")
			return
		}
		if room.Phase != PhaseInProgress {
			c.JSON(http.StatusOK, map[string]string{})
			return
		}
		playerID := c.MustGet("id").(string)
		var seat int
		var concealed []string
		for i, id := range room.Players {
			if id == playerID {
				seat = i
				concealed = room.Round.Hands[i].Concealed
				c.JSON(http.StatusOK, map[string]interface{}{
					"seat":      seat,
					"concealed": concealed,
				})
				return
			}
		}
		c.JSON(http.StatusOK, map[string]string{})
	})
	r.Run("localhost:8080")
}
