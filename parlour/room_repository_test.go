package parlour

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

var (
	conn *pgx.Conn
)

func getTx() pgx.Tx {
	if conn == nil {
		var err error
		conn, err = pgx.Connect(context.Background(), "postgres://localhost/parlour_test?sslmode=disable")
		if err != nil {
			panic(err)
		}
	}
	tx, err := conn.Begin(context.Background())
	if err != nil {
		panic(err)
	}
	return tx
}

func TestPostgresRoomRepository_Save(t *testing.T) {
	t.Run("saves room", func(t *testing.T) {
		tx := getTx()
		defer tx.Rollback(context.Background())

		repo := NewPostgresRoomRepository(tx)
		room := &Room{
			ID: "ABCD",
		}
		err := repo.Save(room)
		assert.NoError(t, err)

		room.Phase = PhaseInProgress
		err = repo.Save(room)
		assert.NoError(t, err)

		got, err := repo.Get("ABCD")
		assert.NoError(t, err)
		assert.Equal(t, PhaseInProgress, got.Phase)
	})
	t.Run("sets room ID if unset", func(t *testing.T) {
		tx := getTx()
		defer tx.Rollback(context.Background())

		repo := NewPostgresRoomRepository(tx)
		room := NewRoom(Player{
			ID:   "iPRk13H8j/MHaP3vhCjnAg",
			Name: "Jiayu",
		})
		err := repo.Save(room)
		assert.NoError(t, err)

		got, err := repo.Get(room.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, got.ID)
	})
	t.Run("generates new ID on conflict", func(t *testing.T) {
		tx := getTx()
		defer tx.Rollback(context.Background())

		repo := NewPostgresRoomRepository(tx)
		room := NewRoom(Player{
			ID:   "iPRk13H8j/MHaP3vhCjnAg",
			Name: "Jiayu",
		})
		err := repo.Save(room)
		assert.NoError(t, err)
		assert.NotEmpty(t, room.ID)
	})
}

func TestPostgresRoomRepository_Get(t *testing.T) {
	tx := getTx()
	defer tx.Rollback(context.Background())

	repo := NewPostgresRoomRepository(tx)
	room := NewRoom(Player{
		ID:   "iPRk13H8j/MHaP3vhCjnAg",
		Name: "Jiayu",
	})
	err := repo.Save(room)
	if err != nil {
		panic(err)
	}

	got, err := repo.Get(room.ID)
	assert.NoError(t, err)
	assert.Equal(t, room, got)
}
