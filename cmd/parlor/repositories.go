package main

import (
	"context"
	crypto "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/go-redis/redis/v8"
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
	return r.rooms[id], nil
}

type RedisRoomRepository struct {
	sync.Mutex
	client *redis.Client
	cache  map[string]*Room
}

func NewRedisRoomRepository(client *redis.Client) *RedisRoomRepository {
	return &RedisRoomRepository{
		client: client,
		cache:  map[string]*Room{},
	}
}

func (r *RedisRoomRepository) key(id string) string {
	return "room:" + id
}

func (r *RedisRoomRepository) Insert(room *Room) (string, error) {
	r.Lock()
	defer r.Unlock()
	data, err := room.MarshalBinary()
	if err != nil {
		return "", err
	}
	for {
		id := newRoomID()
		result, err := r.client.SetNX(context.Background(), r.key(id), data, 0).Result()
		if err != nil {
			return "", err
		}
		if !result {
			continue
		}
		r.cache[id] = room
		return id, nil
	}
}

func (r *RedisRoomRepository) Save(room *Room) error {
	r.Lock()
	defer r.Unlock()
	data, err := room.MarshalBinary()
	if err != nil {
		return err
	}
	r.cache[room.ID] = room
	_, err = r.client.Set(context.Background(), r.key(room.ID), data, 0).Result()
	return err
}

func (r *RedisRoomRepository) Get(id string) (*Room, error) {
	r.Lock()
	defer r.Unlock()
	if room, ok := r.cache[id]; ok {
		return room, nil
	}
	data, err := r.client.Get(context.Background(), r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var room Room
	err = room.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// NamePrefix controls the prefix of the random name which would be generated on insert if Name was empty.
	NamePrefix string `json:"-"`
}

func newPlayerID() string {
	b := make([]byte, 8)
	crypto.Read(b)
	return hex.EncodeToString(b)
}

func randomName(prefix string) string {
	return fmt.Sprintf("%s#%04d", prefix, rand.Intn(1000))
}

type PlayerRepository interface {
	// Insert creates a new player. Implementations should generate the player ID and name if empty.
	Insert(player *Player) error
	Get(id string) (Player, error)
}

type InMemoryPlayerRepository struct {
	sync.RWMutex
	players map[string]Player
	names   map[string]struct{}
}

func (r *InMemoryPlayerRepository) Insert(player *Player) error {
	r.Lock()
	defer r.Unlock()
	if player.ID != "" {
		if _, exists := r.players[player.ID]; exists {
			return errors.New("id already exists")
		}
	} else {
		for {
			id := newPlayerID()
			if _, exists := r.players[id]; !exists {
				player.ID = id
				break
			}
		}
	}
	if player.Name != "" {
		if _, exists := r.names[player.Name]; exists {
			return errors.New("name already exists")
		}
	} else {
		for {
			prefix := "anon"
			if player.NamePrefix != "" {
				prefix = player.NamePrefix
			}
			name := randomName(prefix)
			if _, exists := r.names[name]; !exists {
				player.Name = name
				break
			}
		}
	}
	r.players[player.ID] = *player
	r.names[player.Name] = struct{}{}
	return nil
}

func (r *InMemoryPlayerRepository) Get(id string) (Player, error) {
	r.RLock()
	defer r.RUnlock()
	player, ok := r.players[id]
	if !ok {
		return player, ErrNotFound
	}
	return player, nil
}

func NewInMemoryPlayerRepository() *InMemoryPlayerRepository {
	return &InMemoryPlayerRepository{
		players: map[string]Player{},
		names:   map[string]struct{}{},
	}
}
