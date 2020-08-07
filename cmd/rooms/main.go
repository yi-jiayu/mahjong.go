package main

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
)

var (
	names = make(map[string]string)
	rooms = make(map[string][]string)
)

func newSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func newRoom(players ...string) string {
	charset := "ABCDEFGHIJKLMNOOPQRSTUVWXYZ"
	for {
		id := fmt.Sprintf("%c%c%c%c", charset[rand.Intn(len(charset))], charset[rand.Intn(len(charset))], charset[rand.Intn(len(charset))], charset[rand.Intn(len(charset))])
		if _, ok := rooms[id]; !ok {
			rooms[id] = players
			return id
		}
	}
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
			for {
				name := fmt.Sprintf("anon#%04d", rand.Intn(1000))
				if _, ok := names[id]; !ok {
					names[id] = name
					break
				}
			}
			session.Save()
			c.Set("id", id)
		}
		c.Next()
	})
	r.GET("/self", func(c *gin.Context) {
		playerID := c.MustGet("id").(string)
		name := names[playerID]
		c.JSON(http.StatusOK, map[string]string{
			"id":   playerID,
			"name": name,
		})
	})
	r.GET("/rooms", func(c *gin.Context) {
		var resp strings.Builder
		for roomID, playerIDs := range rooms {
			playerNames := make([]string, len(playerIDs))
			for i, p := range playerIDs {
				playerNames[i] = names[p]
			}
			resp.WriteString(fmt.Sprintf("%s\t%s\n", roomID, strings.Join(playerNames, ", ")))
		}
		c.String(http.StatusOK, "%s", resp.String())
	})
	r.POST("/rooms", func(c *gin.Context) {
		id := c.MustGet("id").(string)
		roomID := newRoom(id)
		c.JSON(http.StatusOK, map[string]string{
			"room_id": roomID,
		})
	})
	r.POST("/rooms/:id/players", func(c *gin.Context) {
		roomID := strings.ToUpper(c.Param("id"))
		playerID := c.MustGet("id").(string)
		room, ok := rooms[roomID]
		if !ok {
			c.String(http.StatusNotFound, "Not Found")
			return
		}
		if len(room) == 4 {
			c.String(http.StatusBadRequest, "Room Full")
			return
		}
		for _, p := range room {
			if p == playerID {
				c.String(http.StatusBadRequest, "Already Joined")
				return
			}
		}
		rooms[roomID] = append(room, playerID)
	})
	r.Run()
}
