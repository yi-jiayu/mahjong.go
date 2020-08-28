package parlour

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

// newPlayerID returns opaque string containing n bytes of entropy.
func newPlayerID(n int) string {
	data := make([]byte, n)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

func setPlayerID(c *gin.Context) {
	playerID, _ := c.Cookie("playerID")
	if playerID == "" {
		playerID = newPlayerID(16)
	}
	c.Set("playerID", playerID)
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("playerID", playerID, 86400, "/", "", false, false)
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
		playerID := c.GetString("playerID")
		name, err := getName(c)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		player := Player{
			id:   playerID,
			Name: name,
		}
		room := NewRoom(player)
		err = roomRepository.Save(room)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			fmt.Printf("error saving room: %v", err)
			return
		}
		c.String(http.StatusCreated, room.ID)
	}
}

func joinRoomHandler(roomRepository RoomRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		playerID := c.GetString("playerID")
		roomID := c.Param("roomID")
		room, err := roomRepository.Get(roomID)
		if errors.Is(err, errNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if err != nil {
			c.Status(http.StatusInternalServerError)
			fmt.Printf("error getting room: %v", err)
			return
		}
		name, err := getName(c)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		player := Player{
			id:   playerID,
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
				c.Status(http.StatusInternalServerError)
				fmt.Printf("error getting room: %v", err)
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
		playerID := c.GetString("playerID")
		roomID := c.Param("roomID")
		room, err := roomRepository.Get(roomID)
		if errors.Is(err, errNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if err != nil {
			c.Status(http.StatusInternalServerError)
			fmt.Printf("error getting room: %v", err)
			return
		}
		room.WithLock(func(r *Room) {
			room.removePlayer(playerID)
			err = roomRepository.Save(r)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				fmt.Printf("error getting room: %v", err)
				return
			}
		})
	}
}

func subscribeRoomHandler(roomRepository RoomRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		playerID := c.GetString("playerID")
		roomID := c.Param("roomID")
		room, err := roomRepository.Get(roomID)
		if errors.Is(err, errNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if err != nil {
			c.Status(http.StatusInternalServerError)
			fmt.Printf("error getting room: %v", err)
			return
		}
		ch := make(chan string, 1)
		room.AddClient(playerID, ch)

		notify := c.Request.Context().Done()
		go func() {
			<-notify
			room.RemoveClient(ch)
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

func configure(r *gin.Engine, roomRepository RoomRepository) {
	r.Use(setPlayerID)
	r.POST("/rooms", createRoomHandler(roomRepository))
	r.POST("/rooms/:roomID/players", joinRoomHandler(roomRepository))
	r.DELETE("/rooms/:roomID/players", leaveRoomHandler(roomRepository))
	r.GET("/rooms/:roomID/live", subscribeRoomHandler(roomRepository))
}
