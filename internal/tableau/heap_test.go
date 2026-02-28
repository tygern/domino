package tableau_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tygern/domino/internal/coxeter"
	"github.com/tygern/domino/internal/tableau"
)

func TestHeap_Construction(t *testing.T) {
	tests := []struct {
		perm   []int
		height int
		width  int
	}{
		{[]int{1, 2, 3, 4}, 0, 0},
		{[]int{1, -4, 3, -2}, 3, 4},
		{[]int{1, -4, 3, -2, 5, 6}, 3, 4},
		{[]int{-1, -6, 3, -4, 5, -2}, 5, 6},
		{[]int{2, 1, 3, 4}, 1, 2},
		{[]int{-1, -2, 3, 4}, 1, 2},
	}
	for _, tt := range tests {
		elem, _ := coxeter.NewElement(tt.perm)
		h := tableau.NewHeap(elem)
		assert.Equal(t, tt.height, h.MaxHeight(), "height mismatch for %v", tt.perm)
		assert.Equal(t, tt.width, h.MaxWidth(), "width mismatch for %v", tt.perm)
	}
}

func TestHeap_Size(t *testing.T) {
	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	h := tableau.NewHeap(w)
	assert.Equal(t, w.Length(), h.Size())

	id := coxeter.NewIdentity(4)
	hid := tableau.NewHeap(id)
	assert.Equal(t, 0, hid.Size())
}

func TestHeap_Rank(t *testing.T) {
	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	assert.Equal(t, 4, tableau.NewHeap(w).Rank())
}

func TestHeap_Blocks(t *testing.T) {
	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	h := tableau.NewHeap(w)
	blocks := h.Blocks()
	assert.Len(t, blocks, w.Length())

	blocks[0] = tableau.Domino{}
	assert.NotEqual(t, tableau.Domino{}, h.Blocks()[0])
}

func TestHeap_SingleGenerator_S1(t *testing.T) {
	w, _ := coxeter.NewElement([]int{-2, -1, 3, 4})
	h := tableau.NewHeap(w)
	assert.Equal(t, 1, h.Size())
	assert.Equal(t, 2, h.MaxWidth())
	assert.Equal(t, 1, h.MaxHeight())
}

func TestHeap_Identity(t *testing.T) {
	id := coxeter.NewIdentity(4)
	h := tableau.NewHeap(id)
	assert.Equal(t, 0, h.Size())
	assert.Empty(t, h.Blocks())
}
