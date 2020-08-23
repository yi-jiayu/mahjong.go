package main

import (
	"math/rand"
	"strings"
	"sync"
)

func newRoomID() string {
	charset := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	return string([]byte{
		charset[rand.Intn(len(charset))],
		charset[rand.Intn(len(charset))],
		charset[rand.Intn(len(charset))],
		charset[rand.Intn(len(charset))],
	})
}

type RoomRepository interface {
	Insert(room *Room) (string, error)
	Save(room *Room) error
	Get(id string) (*Room, error)
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

func (r *InMemoryRoomRepository) Insert(room *Room) (string, error) {
	r.Lock()
	defer r.Unlock()
	for {
		id := newRoomID()
		if _, ok := r.rooms[id]; ok {
			continue
		}
		room.ID = id
		r.rooms[id] = room
		return id, nil
	}
}

func (r *InMemoryRoomRepository) Save(room *Room) error {
	r.Lock()
	defer r.Unlock()
	r.rooms[room.ID] = room
	return nil
}

func (r *InMemoryRoomRepository) Get(id string) (*Room, error) {
	r.RLock()
	defer r.RUnlock()
	id = strings.ToUpper(id)
	room, ok := r.rooms[id]
	if !ok {
		return nil, ErrNotFound
	}
	return room, nil
}
