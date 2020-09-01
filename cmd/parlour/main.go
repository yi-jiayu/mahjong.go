package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gin-contrib/sessions/cookie"
	"github.com/yi-jiayu/mahjong.go/parlour"
)

var host, port string

func init() {
	flag.StringVar(&host, "host", "localhost", "host to listen on")
	flag.StringVar(&port, "port", "8080", "port to listen on")

	rand.Seed(time.Now().UnixNano())
}

func getKey(name string) []byte {
	authKey, err := base64.StdEncoding.DecodeString(os.Getenv(name))
	if err != nil {
		fmt.Printf("error decoding %s: %v\n", name, err)
		os.Exit(1)
	}
	if len(authKey) < 32 {
		fmt.Printf("error: %s should decode to 32 bytes of data\n", name)
		os.Exit(1)
	}
	return authKey
}

func main() {
	roomRepository := parlour.NewInMemoryRoomRepository()
	authKey := getKey("PARLOUR_SESSION_AUTH_KEY")
	encKey := getKey("PARLOUR_SESSION_ENC_KEY")
	store := cookie.NewStore(authKey, encKey)
	p := parlour.Parlour{
		RoomRepository: roomRepository,
		SessionStore:   store,
	}
	err := p.Run(host + ":" + port)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
