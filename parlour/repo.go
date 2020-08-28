package parlour

import (
	"errors"
	"math/rand"
	"strings"
	"sync"
)

var errNotFound = errors.New("not found")

type RoomRepository interface {
	Save(room *Room) error
	Get(id string) (*Room, error)
}

func newRoomID() string {
	charset := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	return string([]byte{
		charset[rand.Intn(len(charset))],
		charset[rand.Intn(len(charset))],
		charset[rand.Intn(len(charset))],
		charset[rand.Intn(len(charset))],
	})
}

type InMemoryRoomRepository struct {
	sync.RWMutex
	rooms map[string]*Room
}

func NewInMemoryRoomRepository() *InMemoryRoomRepository {
	return &InMemoryRoomRepository{
		rooms: map[string]*Room{},
	}
}

func (r *InMemoryRoomRepository) Save(room *Room) error {
	r.Lock()
	defer r.Unlock()
	for room.ID == "" {
		id := newRoomID()
		if _, exists := r.rooms[id]; exists {
			continue
		}
		room.ID = id
	}
	r.rooms[room.ID] = room
	return nil
}

func (r *InMemoryRoomRepository) Get(id string) (*Room, error) {
	r.RLock()
	defer r.RUnlock()
	id = strings.ToUpper(id)
	room, ok := r.rooms[id]
	if !ok {
		return nil, errNotFound
	}
	return room, nil
}
