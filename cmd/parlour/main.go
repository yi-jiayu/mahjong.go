package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/yi-jiayu/mahjong.go/parlour"
)

var host, port string

func init() {
	flag.StringVar(&host, "host", "localhost", "host to listen on")
	flag.StringVar(&port, "port", "8080", "port to listen on")

	rand.Seed(time.Now().UnixNano())
}

func main() {
	roomRepository := parlour.NewInMemoryRoomRepository()
	p := parlour.Parlour{RoomRepository: roomRepository}
	err := p.Run(host + ":" + port)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
