package main

import (
	"context"
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
	Save(id string, room *Room) error
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
		r.rooms[id] = room
		return id, nil
	}
}

func (r *InMemoryRoomRepository) Save(id string, room *Room) error {
	r.Lock()
	defer r.Unlock()
	r.rooms[id] = room
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

func (r *RedisRoomRepository) Save(id string, room *Room) error {
	r.Lock()
	defer r.Unlock()
	data, err := room.MarshalBinary()
	if err != nil {
		return err
	}
	r.cache[id] = room
	_, err = r.client.Set(context.Background(), r.key(id), data, 0).Result()
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
	Name string
}

func newPlayerName() string {
	return fmt.Sprintf("anon#%04d", rand.Intn(1000))
}

type PlayerRepository interface {
	Insert(id string, player Player) error
	Get(id string) (Player, error)
}

type RedisPlayerRepository struct {
	client *redis.Client
}

func (r *RedisPlayerRepository) key(id string) string {
	return "player:" + id
}

func (r *RedisPlayerRepository) Insert(id string, player Player) error {
	panic("implement me")
}

func (r *RedisPlayerRepository) Get(id string) (Player, error) {
	panic("implement me")
}
