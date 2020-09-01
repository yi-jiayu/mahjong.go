package parlour

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Parlour struct {
	RoomRepository RoomRepository
	SessionStore   sessions.Store
}

func (p Parlour) Run(addr string) error {
	r := gin.Default()
	p.configure(r)
	return r.Run(addr)
}
