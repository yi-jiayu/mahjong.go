package parlour

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yi-jiayu/mahjong.go"
)

const (
	KeySessionName = "session"
	KeyPlayerID    = "playerID"
	KeyRoom        = "room"
)

var sessionOptions = sessions.Options{
	Path:     "/",
	SameSite: http.SameSiteStrictMode,
	MaxAge:   2592000, // 30 days in seconds
}

// newPlayerID returns opaque string containing n bytes of entropy.
func newPlayerID(n int) string {
	data := make([]byte, n)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

func setPlayerIDMiddleware(c *gin.Context) {
	session := sessions.Default(c)
	session.Options(sessionOptions)
	var playerID string
	v := session.Get(KeyPlayerID)
	if v == nil {
		playerID = newPlayerID(16)
	} else if p, ok := v.(string); !ok {
		playerID = newPlayerID(16)
	} else {
		playerID = p
	}
	c.Set(KeyPlayerID, playerID)
	session.Set(KeyPlayerID, playerID)
	err := session.Save()
	if err != nil {
		fmt.Printf("error saving session: %v\n", err)
	}
	c.Next()
}

func getName(c *gin.Context) (string, error) {
	name := c.PostForm("name")
	if name == "" {
		return "", errors.New("name is required")
	}
	if ok, _ := regexp.MatchString("[0-9A-Za-z ]+", name); !ok {
		return "", errors.New("name is invalid")
	}
	return name, nil
}

func createRoomHandler(roomRepository RoomRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		playerID := c.GetString(KeyPlayerID)
		name, err := getName(c)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		player := Player{
			ID:   playerID,
			Name: name,
		}
		room := NewRoom(player)
		err = roomRepository.Save(room)
		if err != nil {
			fmt.Printf("error saving room: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		metricOpenRooms.Add(1)
		c.String(http.StatusCreated, room.ID)
	}
}

func joinRoomHandler(roomRepository RoomRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		playerID := c.GetString(KeyPlayerID)
		room := c.MustGet(KeyRoom).(*Room)
		name, err := getName(c)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		player := Player{
			ID:   playerID,
			Name: name,
		}
		room.WithLock(func(r *Room) {
			err = room.addPlayer(player)
			if err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return
			}
			err = roomRepository.Save(r)
			if err != nil {
				fmt.Printf("error saving room: %v", err)
				c.Status(http.StatusInternalServerError)
				return
			}
		})
		if err != nil {
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func leaveRoomHandler(roomRepository RoomRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		playerID := c.GetString(KeyPlayerID)
		room := c.MustGet(KeyRoom).(*Room)
		room.WithLock(func(r *Room) {
			room.removePlayer(playerID)
			err := roomRepository.Save(r)
			if err != nil {
				fmt.Printf("error saving room: %v", err)
				c.Status(http.StatusInternalServerError)
				return
			}
		})
	}
}

func subscribeRoomHandler(_ RoomRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		playerID := c.GetString(KeyPlayerID)
		room := c.MustGet(KeyRoom).(*Room)
		ch := make(chan string, 1)
		room.AddClient(playerID, ch)
		metricRoomSubscriptions.Add(1)

		notify := c.Request.Context().Done()
		go func() {
			<-notify
			room.RemoveClient(ch)
			metricRoomSubscriptions.Add(-1)
		}()

		c.Stream(func(w io.Writer) bool {
			if update, ok := <-ch; ok {
				c.SSEvent("", update)
				return true
			}
			return false
		})
	}
}

func roomActionsHandler(roomRepository RoomRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		playerID := c.GetString(KeyPlayerID)
		room := c.MustGet(KeyRoom).(*Room)
		var action Action
		err := c.ShouldBindJSON(&action)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		room.WithLock(func(r *Room) {
			err = room.reduce(playerID, action)
			if err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return
			}
			err = roomRepository.Save(r)
			if err != nil {
				fmt.Printf("error saving room: %v", err)
				c.Status(http.StatusInternalServerError)
				return
			}
		})
		if err != nil {
			return
		}
	}
}

func setConcealedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		room := c.MustGet(KeyRoom).(*Room)
		if room.Round == nil {
			c.String(http.StatusBadRequest, "round not started")
			return
		}
		seat, err := strconv.Atoi(c.Param("seat"))
		if err != nil {
			c.String(http.StatusBadRequest, "invalid seat")
			return
		}
		if seat < 0 || 3 < seat {
			c.String(http.StatusBadRequest, "invalid seat")
			return
		}
		var tiles mahjong.TileBag
		err = c.ShouldBindJSON(&tiles)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		room.Round.Hands[seat].Concealed = tiles
	}
}

func prependWallHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		room := c.MustGet(KeyRoom).(*Room)
		if room.Round == nil {
			c.String(http.StatusBadRequest, "round not started")
			return
		}
		tile := c.PostForm("tile")
		if tile == "" {
			c.String(http.StatusBadRequest, "tile is required")
			return
		}
		room.Round.Wall = append([]mahjong.Tile{mahjong.Tile(tile)}, room.Round.Wall...)
	}
}

var botNames = [4]string{"Barty Bot", "Francisco Bot", "Lupe Bot", "Mordecai Bot"}

func addBotHandler(_ RoomRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		playerID := c.GetString(KeyPlayerID)
		room := c.MustGet(KeyRoom).(*Room)
		room.WithLock(func(r *Room) {
			if r.seat(playerID) == -1 {
				c.String(http.StatusForbidden, "not in room")
				return
			}
			if len(r.Players) == 4 {
				c.String(http.StatusBadRequest, "room full")
				return
			}
			botID := newPlayerID(16)
			r.Players = append(r.Players, Player{
				ID:   botID,
				Name: botNames[len(r.Players)],
			})
			r.broadcast()
			ch := make(chan string, 1)
			r.clients[ch] = botID
			bot := Bot{
				ID:      botID,
				Room:    room,
				Updates: ch,
			}
			go bot.Start()
		})
	}
}

func setRoomMiddleware(roomRepository RoomRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID := c.Param("roomID")
		room, err := roomRepository.Get(roomID)
		if errors.Is(err, errNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		if err != nil {
			fmt.Printf("error getting room: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Set("room", room)
		c.Next()
	}
}

func (p Parlour) configure(r *gin.Engine) {
	r.Use(sessions.Sessions(KeySessionName, p.SessionStore))
	r.Use(setPlayerIDMiddleware)
	r.POST("/rooms", createRoomHandler(p.RoomRepository))
	room := r.Group("/rooms/:roomID")
	room.Use(setRoomMiddleware(p.RoomRepository))
	{
		room.POST("/players", joinRoomHandler(p.RoomRepository))
		room.DELETE("/players", leaveRoomHandler(p.RoomRepository))
		room.GET("/live", subscribeRoomHandler(p.RoomRepository))
		room.POST("/actions", roomActionsHandler(p.RoomRepository))
		room.POST("/bots", addBotHandler(p.RoomRepository))
		if gin.IsDebugging() {
			room.PUT("/round/hands/:seat/concealed", setConcealedHandler())
			room.POST("/round/wall", prependWallHandler())
		}
	}
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
