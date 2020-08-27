package main

import (
	"flag"
	"fmt"

	"github.com/yi-jiayu/mahjong.go/parlour"
)

var host, port string

func init() {
	flag.StringVar(&host, "host", "localhost", "host to listen on")
	flag.StringVar(&port, "port", "8080", "port to listen on")
}

func main() {
	roomRepository := parlour.NewInMemoryRoomRepository()
	p := parlour.Parlour{RoomRepository: roomRepository}
	err := p.Run(host + ":" + port)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
