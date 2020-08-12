package main

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

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

var (
	playerRegistry = &PlayerRegistry{
		Names: map[string]string{},
	}
	roomRepository RoomRepository = NewInMemoryRoomRepository()
)

func newSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	r := gin.Default()
	store := memstore.NewStore([]byte("secret"))
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
		room := NewRoom(id)
		roomID, _ := roomRepository.Insert(room)
		c.JSON(http.StatusOK, map[string]string{
			"room_id": roomID,
		})
	})
	r.GET("/rooms/:id/live", func(c *gin.Context) {
		roomID := strings.ToUpper(c.Param("id"))
		room, _ := roomRepository.Get(roomID)
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
		go room.AddClient(ch)

		// Remove this client from the map of connected clients
		// when this handler exits.
		defer room.RemoveClient(ch)

		// Listen to connection close and un-register c
		notify := c.Request.Context().Done()
		go func() {
			<-notify
			room.RemoveClient(ch)
		}()

		for {
			game := <-ch
			fmt.Fprintf(c.Writer, "data: %s\n\n", game)
			flusher.Flush()
		}
	})
	r.POST("/rooms/:id/players", func(c *gin.Context) {
		roomID := strings.ToUpper(c.Param("id"))
		room, _ := roomRepository.Get(roomID)
		if room == nil {
			c.String(http.StatusNotFound, "Not Found")
			return
		}
		playerID := c.MustGet("id").(string)
		err := room.AddPlayer(playerID)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
	})
	r.POST("/rooms/:id/actions", func(c *gin.Context) {
		roomID := strings.ToUpper(c.Param("id"))
		room, _ := roomRepository.Get(roomID)
		if room == nil {
			c.String(http.StatusNotFound, "Not Found")
			return
		}
		playerID := c.MustGet("id").(string)
		var action Action
		if err := c.ShouldBindJSON(&action); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		if err := room.HandleAction(playerID, action); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
	})
	r.GET("/rooms/:id/self", func(c *gin.Context) {
		roomID := strings.ToUpper(c.Param("id"))
		room, _ := roomRepository.Get(roomID)
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
		var concealed []mahjong.Tile
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
	if gin.IsDebugging() {
		r.POST("/debug/rooms/:id/reshuffle", func(c *gin.Context) {
			roomID := strings.ToUpper(c.Param("id"))
			room, _ := roomRepository.Get(roomID)
			if room == nil {
				c.String(http.StatusNotFound, "Not Found")
				return
			}
			if room.Phase != PhaseInProgress {
				c.String(http.StatusBadRequest, "not in progress")
				return
			}
			room.Round = mahjong.NewRound(rand.Int63())
			room.broadcast()
		})
		r.GET("/debug/rooms/:id/wall", func(c *gin.Context) {
			roomID := strings.ToUpper(c.Param("id"))
			room, _ := roomRepository.Get(roomID)
			if room == nil {
				c.String(http.StatusNotFound, "Not Found")
				return
			}
			if room.Phase != PhaseInProgress {
				c.String(http.StatusBadRequest, "not in progress")
				return
			}
			c.JSON(http.StatusOK, room.Round.Wall)
		})
		r.POST("/debug/rooms/:id/wall", func(c *gin.Context) {
			roomID := strings.ToUpper(c.Param("id"))
			room, _ := roomRepository.Get(roomID)
			if room == nil {
				c.String(http.StatusNotFound, "Not Found")
				return
			}
			if room.Phase != PhaseInProgress {
				c.String(http.StatusBadRequest, "not in progress")
				return
			}
			tile := c.Query("tile")
			if tile == "" {
				c.String(http.StatusBadRequest, "tile not provided")
				return
			}
			room.Round.Wall = append([]mahjong.Tile{mahjong.Tile(tile)}, room.Round.Wall...)
			room.broadcast()
		})
		r.POST("/debug/rooms/:id/round/hands/:seat/concealed", func(c *gin.Context) {
			roomID := strings.ToUpper(c.Param("id"))
			room, _ := roomRepository.Get(roomID)
			if room == nil {
				c.String(http.StatusNotFound, "Not Found")
				return
			}
			if room.Phase != PhaseInProgress {
				c.String(http.StatusBadRequest, "not in progress")
				return
			}
			seat, err := strconv.Atoi(c.Param("seat"))
			if err != nil {
				c.String(http.StatusBadRequest, "invalid seat")
				return
			}
			var tiles []mahjong.Tile
			err = c.ShouldBindJSON(&tiles)
			if err != nil {
				c.Status(http.StatusBadRequest)
				return
			}
			room.Round.Hands[seat].Concealed = tiles
		})
	}
	r.Run("localhost:8080")
}
