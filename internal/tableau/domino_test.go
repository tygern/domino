package tableau_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tygern/domino/internal/tableau"
)

func TestDomino_SecondBlock(t *testing.T) {
	vert := tableau.Domino{Label: 1, Col: 3, Row: 5, IsVertical: true}
	assert.Equal(t, 3, vert.SecondCol())
	assert.Equal(t, 6, vert.SecondRow())

	horiz := tableau.Domino{Label: 2, Col: 3, Row: 5, IsVertical: false}
	assert.Equal(t, 4, horiz.SecondCol())
	assert.Equal(t, 5, horiz.SecondRow())
}

func TestDomino_MoveTo(t *testing.T) {
	d := tableau.Domino{Label: 3, Col: 1, Row: 1, IsVertical: true}
	moved := d.MoveTo(5, 7)

	assert.Equal(t, 3, moved.Label)
	assert.Equal(t, 5, moved.Col)
	assert.Equal(t, 7, moved.Row)
	assert.True(t, moved.IsVertical)
}

func TestDomino_Flip(t *testing.T) {
	d := tableau.Domino{Label: 4, Col: 2, Row: 3, IsVertical: true}
	flipped := d.Flip()

	assert.Equal(t, 4, flipped.Label)
	assert.Equal(t, 2, flipped.Col)
	assert.Equal(t, 3, flipped.Row)
	assert.False(t, flipped.IsVertical)

	assert.True(t, d.Flip().Flip().Equal(d))
}

func TestDomino_Equal(t *testing.T) {
	a := tableau.Domino{Label: 1, Col: 2, Row: 3, IsVertical: true}
	b := tableau.Domino{Label: 1, Col: 2, Row: 3, IsVertical: true}
	c := tableau.Domino{Label: 1, Col: 2, Row: 3, IsVertical: false}

	assert.True(t, a.Equal(b))
	assert.False(t, a.Equal(c))
}
