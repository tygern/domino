package tableau_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tygern/domino/internal/coxeter"
	"github.com/tygern/domino/internal/tableau"
)

func TestTableau_Construction(t *testing.T) {
	w4, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	v4, _ := coxeter.NewElement([]int{1, 3, -4, -2})
	assert.True(t, tableau.New(w4).Equal(tableau.New(v4)))

	u4, _ := coxeter.NewElement([]int{1, 2, 3, 4})
	assert.False(t, tableau.New(w4).Equal(tableau.New(u4)))

	w46, _ := coxeter.NewElement([]int{1, -4, 3, -2, 5, 6})
	assert.False(t, tableau.New(w4).Equal(tableau.New(w46)))

	w6, _ := coxeter.NewElement([]int{-1, -6, 3, -4, 5, -2})
	v6, _ := coxeter.NewElement([]int{-6, -1, 3, -4, -2, 5})
	assert.True(t, tableau.New(w6).Equal(tableau.New(v6)))
}

func TestTableau_Size(t *testing.T) {
	w4, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	assert.Equal(t, 4, tableau.New(w4).Size())

	w6, _ := coxeter.NewElement([]int{-1, -6, 3, -4, 5, -2})
	assert.Equal(t, 6, tableau.New(w6).Size())

	w46, _ := coxeter.NewElement([]int{1, -4, 3, -2, 5, 6})
	assert.Equal(t, 6, tableau.New(w46).Size())
}

func TestTableau_Dimensions(t *testing.T) {
	w4, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	tw4 := tableau.New(w4)
	assert.Equal(t, 4, tw4.MaxWidth())
	assert.Equal(t, 3, tw4.MaxHeight())

	w6, _ := coxeter.NewElement([]int{-1, -6, 3, -4, 5, -2})
	tw6 := tableau.New(w6)
	assert.Equal(t, 5, tw6.MaxWidth())
	assert.Equal(t, 4, tw6.MaxHeight())

	w46, _ := coxeter.NewElement([]int{1, -4, 3, -2, 5, 6})
	tw46 := tableau.New(w46)
	assert.Equal(t, 8, tw46.MaxWidth())
	assert.Equal(t, 3, tw46.MaxHeight())
}

func TestTableau_GetDomino(t *testing.T) {
	w4, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	tw4 := tableau.New(w4)

	d, ok := tw4.GetDomino(4)
	assert.True(t, ok)
	assert.Equal(t, 2, d.Col)
	assert.Equal(t, 2, d.Row)
	assert.True(t, d.IsVertical)

	w6, _ := coxeter.NewElement([]int{-1, -6, 3, -4, 5, -2})
	tw6 := tableau.New(w6)

	d5, ok := tw6.GetDomino(5)
	assert.True(t, ok)
	assert.Equal(t, 4, d5.Col)
	assert.Equal(t, 1, d5.Row)
	assert.False(t, d5.IsVertical)
}

func TestTableau_Rank(t *testing.T) {
	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	assert.Equal(t, 4, tableau.New(w).Rank())

	w6, _ := coxeter.NewElement([]int{-1, -6, 3, -4, 5, -2})
	assert.Equal(t, 6, tableau.New(w6).Rank())
}

func TestTableau_Dominoes(t *testing.T) {
	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	doms := tableau.New(w).Dominoes()
	assert.Len(t, doms, 4)

	labels := make([]int, len(doms))
	for i, d := range doms {
		labels[i] = d.Label
	}
	assert.Equal(t, []int{1, 2, 3, 4}, labels)
}

func TestTableau_GetDomino_Invalid(t *testing.T) {
	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	tw := tableau.New(w)

	_, ok := tw.GetDomino(0)
	assert.False(t, ok)

	_, ok = tw.GetDomino(5)
	assert.False(t, ok)
}

func TestTableau_Equal_DifferentDominoes(t *testing.T) {
	a, _ := coxeter.NewElement([]int{1, 2, 3, 4})
	b, _ := coxeter.NewElement([]int{2, 1, 3, 4})
	assert.False(t, tableau.New(a).Equal(tableau.New(b)))
}

func TestTableau_Identity(t *testing.T) {
	id := coxeter.NewIdentity(4)
	tid := tableau.New(id)

	assert.Equal(t, 4, tid.Size())
	for i := 1; i <= 4; i++ {
		d, ok := tid.GetDomino(i)
		assert.True(t, ok)
		assert.False(t, d.IsVertical)
		assert.Equal(t, 1, d.Row)
	}
}

func TestTableau_RightAndLeft(t *testing.T) {
	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	left := tableau.LeftTableau(w)
	rightOfInverse := tableau.RightTableau(w.Inverse())
	assert.True(t, left.Equal(rightOfInverse))
}

func TestTableau_OutOfOrderLabels(t *testing.T) {
	// [3,1,2,4]: label 3 is inserted first, then label 1 overwrites its grid cells.
	// When label 3 is shuffled, removeDomino encounters a first cell already overwritten.
	elem, _ := coxeter.NewElement([]int{3, 1, 2, 4})
	tab := tableau.New(elem)
	assert.Equal(t, 4, tab.Size())

	for i := 1; i <= 4; i++ {
		_, ok := tab.GetDomino(i)
		assert.True(t, ok, "domino %d should be present", i)
	}
}

func TestTableau_Exhaustive_D4(t *testing.T) {
	all := coxeter.AllElements(4)
	assert.Len(t, all, 192)

	for _, elem := range all {
		tab := tableau.New(elem)
		assert.Equal(t, 4, tab.Size(), "element %v", elem)
		assert.True(t, tab.MaxWidth() > 0, "element %v", elem)
		assert.True(t, tab.MaxHeight() > 0, "element %v", elem)
	}

	for _, a := range all {
		for _, b := range all {
			ta := tableau.RightTableau(a)
			tb := tableau.RightTableau(b)
			if ta.Equal(tb) {
				assert.Equal(t, a.LeftDescentSet(), b.LeftDescentSet(),
					"same right tableau but different left descents: %v vs %v", a, b)
			}

			la := tableau.LeftTableau(a)
			lb := tableau.LeftTableau(b)
			if la.Equal(lb) {
				assert.Equal(t, a.RightDescentSet(), b.RightDescentSet(),
					"same left tableau but different right descents: %v vs %v", a, b)
			}
		}
	}
}
