package coxeter_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tygern/domino/internal/coxeter"
)

func TestAllElements(t *testing.T) {
	elems3 := coxeter.AllElements(3)
	assert.Len(t, elems3, 24)

	elems4 := coxeter.AllElements(4)
	assert.Len(t, elems4, 192)
}

func TestBadElements_D3(t *testing.T) {
	bad := coxeter.BadElements(3)
	assert.Empty(t, bad)
}

func TestBadElements_D4(t *testing.T) {
	bad := coxeter.BadElements(4)

	for _, elem := range bad {
		assert.True(t, elem.IsBad(), "returned non-bad element: %v", elem)
	}

	known, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	found := false
	for _, elem := range bad {
		if elem.Equal(known) {
			found = true
			break
		}
	}
	assert.True(t, found, "known bad element [1, -4, 3, -2] not found")

	assert.True(t, len(bad) > 0, "D_4 should have bad elements")
}

func TestBadElements_D5(t *testing.T) {
	bad := coxeter.BadElements(5)

	for _, elem := range bad {
		assert.True(t, elem.IsBad(), "returned non-bad element: %v", elem)
	}

	assert.True(t, len(bad) > 0, "D_5 should have bad elements")
}

func TestBadElements_BruteForceComparison_D4(t *testing.T) {
	all := coxeter.AllElements(4)
	var bruteForceBad []string
	for _, e := range all {
		if e.IsBad() {
			bruteForceBad = append(bruteForceBad, e.String())
		}
	}
	sort.Strings(bruteForceBad)

	optimized := coxeter.BadElements(4)
	var optimizedBad []string
	for _, e := range optimized {
		optimizedBad = append(optimizedBad, e.String())
	}
	sort.Strings(optimizedBad)

	assert.Equal(t, bruteForceBad, optimizedBad)
}

func TestBadElements_BruteForceComparison_D5(t *testing.T) {
	all := coxeter.AllElements(5)
	var bruteForceBad []string
	for _, e := range all {
		if e.IsBad() {
			bruteForceBad = append(bruteForceBad, e.String())
		}
	}
	sort.Strings(bruteForceBad)

	optimized := coxeter.BadElements(5)
	var optimizedBad []string
	for _, e := range optimized {
		optimizedBad = append(optimizedBad, e.String())
	}
	sort.Strings(optimizedBad)

	assert.Equal(t, bruteForceBad, optimizedBad)
}

func TestBadElements_BruteForceComparison_D6(t *testing.T) {
	all := coxeter.AllElements(6)
	var bruteForceBad []string
	for _, e := range all {
		if e.IsBad() {
			bruteForceBad = append(bruteForceBad, e.String())
		}
	}
	sort.Strings(bruteForceBad)

	optimized := coxeter.BadElements(6)
	var optimizedBad []string
	for _, e := range optimized {
		optimizedBad = append(optimizedBad, e.String())
	}
	sort.Strings(optimizedBad)

	assert.Equal(t, bruteForceBad, optimizedBad)
}

func TestBadElements_BruteForceComparison_D7(t *testing.T) {
	all := coxeter.AllElements(7)
	var bruteForceBad []string
	for _, e := range all {
		if e.IsBad() {
			bruteForceBad = append(bruteForceBad, e.String())
		}
	}
	sort.Strings(bruteForceBad)

	optimized := coxeter.BadElements(7)
	var optimizedBad []string
	for _, e := range optimized {
		optimizedBad = append(optimizedBad, e.String())
	}
	sort.Strings(optimizedBad)

	assert.Equal(t, bruteForceBad, optimizedBad)
}

func TestBadElements_BruteForceComparison_D8(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping D_8 brute force comparison in short mode")
	}

	all := coxeter.AllElements(8)
	var bruteForceBad []string
	for _, e := range all {
		if e.IsBad() {
			bruteForceBad = append(bruteForceBad, e.String())
		}
	}
	sort.Strings(bruteForceBad)

	optimized := coxeter.BadElements(8)
	var optimizedBad []string
	for _, e := range optimized {
		optimizedBad = append(optimizedBad, e.String())
	}
	sort.Strings(optimizedBad)

	assert.Equal(t, bruteForceBad, optimizedBad)
}

func TestBadElements_KnownBadFromThesis(t *testing.T) {
	known := [][]int{
		{1, -4, 3, -2},
		{-1, -6, 3, -4, 5, -2},
		{1, -8, 3, -6, 5, -4, 7, -2},
	}

	for _, perm := range known {
		elem, err := coxeter.NewElement(perm)
		assert.NoError(t, err)
		assert.True(t, elem.IsBad(), "known bad element %v should be bad", perm)

		bad := coxeter.BadElements(len(perm))
		found := false
		for _, b := range bad {
			if b.Equal(elem) {
				found = true
				break
			}
		}
		assert.True(t, found, "known bad element %v not in BadElements(%d)", perm, len(perm))
	}
}

func TestBadElements_KnownCounts(t *testing.T) {
	expected := map[int]int{
		3: 0,
		4: 1,
		6: 3,
	}

	for rank, count := range expected {
		bad := coxeter.BadElements(rank)
		assert.Len(t, bad, count, "D_%d bad element count", rank)
	}
}

func TestBadElements_AllResultsAreBad(t *testing.T) {
	for rank := 4; rank <= 14; rank++ {
		bad := coxeter.BadElements(rank)
		for _, elem := range bad {
			assert.True(t, elem.IsBad(), "D_%d: returned non-bad element: %v", rank, elem)
		}
		t.Logf("D_%d: %d bad elements, all verified", rank, len(bad))
	}
}
