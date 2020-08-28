package parlour

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoom_AddPlayer(t *testing.T) {
	t.Run("name already taken", func(t *testing.T) {
		r := NewRoom(Player{id: "id1", Name: "player1"})
		err := r.addPlayer(Player{id: "id2", Name: "player1"})
		assert.EqualError(t, err, "name already taken")
	})
	t.Run("room full", func(t *testing.T) {
		r := NewRoom(Player{Name: "player1"})
		_ = r.addPlayer(Player{Name: "player2"})
		_ = r.addPlayer(Player{Name: "player3"})
		_ = r.addPlayer(Player{Name: "player4"})
		err := r.addPlayer(Player{Name: "player5"})
		assert.EqualError(t, err, "room full")
	})
	t.Run("success", func(t *testing.T) {
		r := NewRoom(Player{Name: "player1"})
		err := r.addPlayer(Player{Name: "player2"})
		assert.NoError(t, err)
		assert.Equal(t, []Player{
			{Name: "player1"},
			{Name: "player2"},
		}, r.Players)
	})
}
