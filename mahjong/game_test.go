package mahjong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGame_NextRound(t *testing.T) {
	t.Run("cannot start next round when current round is not finished", func(t *testing.T) {
		r := &Round{Phase: PhaseDiscard}
		g := &Game{Round: r}
		err := g.NextRound(0)
		assert.EqualError(t, err, "current round not finished")
	})
	t.Run("dealer moves on", func(t *testing.T) {
		r := &Round{
			Phase:  PhaseFinished,
			Dealer: 0,
			Result: Result{
				Winner: 2,
			},
		}
		g := &Game{
			Round: r,
		}
		err := g.NextRound(0)
		assert.NoError(t, err)
		assert.Equal(t, 1, g.Round.Dealer)
		assert.Equal(t, []Round{*r}, g.PreviousRounds)
	})
	t.Run("dealer wins and remains dealer", func(t *testing.T) {
		r := &Round{
			Phase:  PhaseFinished,
			Dealer: 2,
			Result: Result{
				Winner: 2,
			},
		}
		g := &Game{
			Round: r,
		}
		err := g.NextRound(0)
		assert.NoError(t, err)
		assert.Equal(t, 2, g.Round.Dealer)
		assert.Equal(t, []Round{*r}, g.PreviousRounds)
	})
	t.Run("change of wind", func(t *testing.T) {
		r := &Round{
			Phase:  PhaseFinished,
			Dealer: 3,
			Wind:   DirectionEast,
			Result: Result{
				Winner: 0,
			},
		}
		g := &Game{
			Round: r,
		}
		err := g.NextRound(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, g.Round.Dealer)
		assert.Equal(t, DirectionSouth, g.Round.Wind)
		assert.Equal(t, []Round{*r}, g.PreviousRounds)
	})
}
