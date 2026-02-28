package tikz_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tygern/domino/internal/coxeter"
	"github.com/tygern/domino/internal/tableau"
	"github.com/tygern/domino/internal/tikz"
)

func TestRenderTableau(t *testing.T) {
	elem, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	tab := tableau.New(elem)
	result := tikz.RenderTableau(tab)

	assert.True(t, strings.Contains(result, "\\begin{tikzpicture}"))
	assert.True(t, strings.Contains(result, "\\end{tikzpicture}"))
	assert.True(t, strings.Contains(result, "\\tikzstyle{ver}"))
	assert.True(t, strings.Contains(result, "\\tikzstyle{hor}"))

	nodeCount := strings.Count(result, "\\node")
	assert.Equal(t, 4, nodeCount)

	assert.True(t, strings.Contains(result, "[hor]"))
	assert.True(t, strings.Contains(result, "[ver]"))
}

func TestRenderTableau_Identity(t *testing.T) {
	id := coxeter.NewIdentity(3)
	tab := tableau.New(id)
	result := tikz.RenderTableau(tab)

	assert.Equal(t, 0, strings.Count(result, "[ver]"))
	assert.Equal(t, 3, strings.Count(result, "[hor]"))
}

func TestRenderHeap(t *testing.T) {
	elem, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	h := tableau.NewHeap(elem)
	result := tikz.RenderHeap(h)

	assert.True(t, strings.Contains(result, "\\begin{tikzpicture}"))
	assert.True(t, strings.Contains(result, "\\end{tikzpicture}"))

	nodeCount := strings.Count(result, "\\node")
	assert.Equal(t, elem.Length(), nodeCount)
}

func TestRenderHeap_PhantomSpacing(t *testing.T) {
	elem, _ := coxeter.NewElement([]int{-1, -2, 3, 4})
	h := tableau.NewHeap(elem)
	result := tikz.RenderHeap(h)

	assert.True(t, strings.Contains(result, "1\\phantom{ 2}"))

	elem2, _ := coxeter.NewElement([]int{1, 3, 2, 4})
	h2 := tableau.NewHeap(elem2)
	result2 := tikz.RenderHeap(h2)

	assert.False(t, strings.Contains(result2, "\\phantom"))
}
