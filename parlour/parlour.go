package parlour

import (
	"github.com/gin-gonic/gin"
)

type Parlour struct {
	RoomRepository RoomRepository
}

func (p Parlour) Run(addr string) error {
	r := gin.Default()
	configure(r, p.RoomRepository)
	return r.Run(addr)
}
