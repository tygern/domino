package coxeter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tygern/domino/internal/coxeter"
)

func TestElement_NewElement(t *testing.T) {
	_, err := coxeter.NewElement([]int{1, 2, 3, 4})
	assert.NoError(t, err)

	_, err = coxeter.NewElement([]int{1, -4, 3, -2})
	assert.NoError(t, err)

	_, err = coxeter.NewElement([]int{-4, 1, 2, 3})
	assert.Error(t, err)

	_, err = coxeter.NewElement([]int{-7, 1, -2, 3})
	assert.Error(t, err)

	_, err = coxeter.NewElement([]int{-4, 0, -2, 3})
	assert.Error(t, err)

	_, err = coxeter.NewElement([]int{2, 1, -2, 3})
	assert.Error(t, err)

	_, err = coxeter.NewElement([]int{})
	assert.Error(t, err)
}

func TestElement_MapsTo(t *testing.T) {
	v, _ := coxeter.NewElement([]int{1, 2, 3, 4})
	assert.Equal(t, 3, v.MapsTo(3))
	assert.Equal(t, -3, v.MapsTo(-3))

	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	assert.Equal(t, -2, w.MapsTo(4))
	assert.Equal(t, 2, w.MapsTo(-4))
	assert.Equal(t, 0, w.MapsTo(5))

	y, _ := coxeter.NewElement([]int{1, 3, -4, -2})
	assert.Equal(t, 4, y.MapsTo(-3))
}

func TestElement_Inverse(t *testing.T) {
	v, _ := coxeter.NewElement([]int{1, 2, 3, 4})
	assert.True(t, v.Inverse().Equal(v))

	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	assert.True(t, w.Inverse().Equal(w))

	y, _ := coxeter.NewElement([]int{1, 3, -4, -2})
	yInv, _ := coxeter.NewElement([]int{1, -4, 2, -3})
	assert.True(t, y.Inverse().Equal(yInv))
	assert.True(t, y.Inverse().Inverse().Equal(y))

	id := coxeter.NewIdentity(4)
	assert.True(t, id.Inverse().Equal(id))
}

func TestElement_RightMultiply(t *testing.T) {
	u, _ := coxeter.NewElement([]int{1, 2, 3, 4})
	v, _ := coxeter.NewElement([]int{1, 2, 3, 4})
	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	y, _ := coxeter.NewElement([]int{1, 3, -4, -2})
	z, _ := coxeter.NewElement([]int{1, 3, 2, 4})

	assert.True(t, u.RightMultiply(v).Equal(u))
	assert.True(t, w.RightMultiply(y).Equal(z))

	yw, _ := coxeter.NewElement([]int{1, 2, -4, -3})
	assert.True(t, y.RightMultiply(w).Equal(yw))

	assert.True(t, w.RightMultiply(w.Inverse()).Equal(coxeter.NewIdentity(4)))

	a, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	b, _ := coxeter.NewElement([]int{1, 3, -4, -2})
	c, _ := coxeter.NewElement([]int{1, 3, 2, 4})
	ab := a.RightMultiply(b)
	bc := b.RightMultiply(c)
	assert.True(t, ab.RightMultiply(c).Equal(a.RightMultiply(bc)))
}

func TestElement_LeftMultiply(t *testing.T) {
	u, _ := coxeter.NewElement([]int{1, 2, 3, 4})
	v, _ := coxeter.NewElement([]int{1, 2, 3, 4})
	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	y, _ := coxeter.NewElement([]int{1, 3, -4, -2})
	z, _ := coxeter.NewElement([]int{1, 3, 2, 4})

	assert.True(t, u.LeftMultiply(v).Equal(u))
	assert.True(t, y.LeftMultiply(w).Equal(z))
}

func TestElement_Length(t *testing.T) {
	tests := []struct {
		perm   []int
		length int
	}{
		{[]int{1, 2, 3, 4}, 0},
		{[]int{1, -4, 3, -2}, 7},
		{[]int{1, 3, 2, 4}, 1},
		{[]int{1, 3, -4, -2}, 8},
	}
	for _, tt := range tests {
		elem, _ := coxeter.NewElement(tt.perm)
		assert.Equal(t, tt.length, elem.Length(), "perm=%v", tt.perm)
	}

	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	assert.Equal(t, w.Length(), w.Inverse().Length())
}

func TestElement_IsRightDescent(t *testing.T) {
	id := coxeter.NewIdentity(4)
	assert.Empty(t, id.RightDescentSet())

	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	assert.Equal(t, []int{1, 2, 4}, w.RightDescentSet())

	z, _ := coxeter.NewElement([]int{1, 3, 2, 4})
	assert.Equal(t, []int{3}, z.RightDescentSet())
}

func TestElement_LeftDescentSet(t *testing.T) {
	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	assert.Equal(t, w.Inverse().RightDescentSet(), w.LeftDescentSet())

	y, _ := coxeter.NewElement([]int{1, 3, -4, -2})
	assert.Equal(t, y.Inverse().RightDescentSet(), y.LeftDescentSet())
}

func TestElement_Equal(t *testing.T) {
	a, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	b, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	assert.True(t, a.Equal(b))

	c, _ := coxeter.NewElement([]int{1, 3, 2, 4})
	assert.False(t, a.Equal(c))

	d, _ := coxeter.NewElement([]int{1, 2, 3})
	assert.False(t, a.Equal(d))
}

func TestElement_IsRightDescent_OutOfRange(t *testing.T) {
	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	assert.False(t, w.IsRightDescent(0))
	assert.False(t, w.IsRightDescent(5))
	assert.False(t, w.IsRightDescent(-1))
}

func TestElement_IsBad(t *testing.T) {
	tests := []struct {
		perm []int
		bad  bool
	}{
		{[]int{1, -4, 3, -2}, true},
		{[]int{-1, -6, 3, -4, 5, -2}, true},
		{[]int{1, -8, 3, -6, 5, -4, 7, -2}, true},
		{[]int{1, 2, 3, 4}, false},
		{[]int{1, 3, -4, -2}, false},
		{[]int{1, 3, 2, 4}, false},
		{[]int{-1, 3, 4, -2}, false},
	}
	for _, tt := range tests {
		elem, err := coxeter.NewElement(tt.perm)
		assert.NoError(t, err)
		assert.Equal(t, tt.bad, elem.IsBad(), "perm=%v", tt.perm)
	}
}
