package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"

	"github.com/yi-jiayu/mahjong.go"
)

var (
	roomRepository RoomRepository = NewInMemoryRoomRepository()
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getPlayerID(sess sessions.Session) string {
	id := sess.Get("id")
	if id != nil {
		return ""
	}
	playerID, ok := id.(string)
	if !ok {
		return ""
	}
	return playerID
}

func main() {
	r := gin.Default()
	store := memstore.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("MJSESSIONID", store))
	r.Use(func(c *gin.Context) {
		session := sessions.Default(c)
		session.Options(sessions.Options{
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		playerID := getPlayerID(session)
		if playerID == "" {
			playerID = newPlayerID()
			session.Set("id", playerID)
		}
		c.Set("id", playerID)
		session.Save()
		c.Next()
	})
	r.POST("/rooms", func(c *gin.Context) {
		playerID := c.GetString("id")
		name := c.PostForm("name")
		if name == "" {
			c.String(http.StatusBadRequest, "id is required")
			return
		}
		room := NewRoom(playerID, name)
		roomID, _ := roomRepository.Insert(room)
		c.JSON(http.StatusOK, map[string]string{
			"room_id": roomID,
		})
	})
	r.GET("/rooms/:id/live", func(c *gin.Context) {
		room, _ := roomRepository.Get(c.Param("id"))
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

		playerID := c.GetString("id")
		ch := make(chan string)
		room.AddClient(playerID, ch)

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
		room, _ := roomRepository.Get(c.Param("id"))
		if room == nil {
			c.String(http.StatusNotFound, "Not Found")
			return
		}
		name := c.PostForm("name")
		if name == "" {
			c.String(http.StatusBadRequest, "id is required")
			return
		}
		playerID := c.GetString("id")
		err := room.AddPlayer(playerID, name)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
	})
	r.POST("/rooms/:id/bots", func(c *gin.Context) {
		room, _ := roomRepository.Get(c.Param("id"))
		if room == nil {
			c.String(http.StatusNotFound, "Not Found")
			return
		}
		playerID := c.GetString("id")
		bot := NewBot(room.ID)
		err := room.AddBot(playerID, bot)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}
		go bot.Run()
	})
	r.POST("/rooms/:id/actions", func(c *gin.Context) {
		room, _ := roomRepository.Get(c.Param("id"))
		if room == nil {
			c.String(http.StatusNotFound, "Not Found")
			return
		}
		playerID := c.GetString("id")
		var action Action
		if err := c.ShouldBindJSON(&action); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		result, err := room.HandleAction(playerID, action)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		c.JSON(http.StatusOK, result)
	})
	if gin.IsDebugging() {
		r.POST("/debug/rooms/:id/reshuffle", func(c *gin.Context) {
			room, _ := roomRepository.Get(c.Param("id"))
			if room == nil {
				c.String(http.StatusNotFound, "Not Found")
				return
			}
			if room.Phase != PhaseInProgress {
				c.String(http.StatusBadRequest, "not in progress")
				return
			}
			room.Round = mahjong.NewRound(rand.Int63())
			room.Round.PongDuration = 2 * time.Second
			room.broadcast()
		})
		r.GET("/debug/rooms/:id/wall", func(c *gin.Context) {
			room, _ := roomRepository.Get(c.Param("id"))
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
			room, _ := roomRepository.Get(c.Param("id"))
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
			room, _ := roomRepository.Get(c.Param("id"))
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
