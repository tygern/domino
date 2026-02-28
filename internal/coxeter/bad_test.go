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
