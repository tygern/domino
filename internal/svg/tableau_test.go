package svg_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tygern/domino/internal/coxeter"
	"github.com/tygern/domino/internal/svg"
	"github.com/tygern/domino/internal/tableau"
)

func TestRenderTableau(t *testing.T) {
	elem, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	tab := tableau.New(elem)
	result := svg.RenderTableau(tab)

	assert.True(t, strings.Contains(result, "<svg"))
	assert.True(t, strings.Contains(result, "</svg>"))

	rectCount := strings.Count(result, "<rect")
	assert.Equal(t, 4, rectCount)

	textCount := strings.Count(result, "<text")
	assert.Equal(t, 4, textCount)

	assert.True(t, strings.Contains(result, `height="80"`))
	assert.True(t, strings.Contains(result, `height="40"`))
}

func TestRenderTableau_Identity(t *testing.T) {
	id := coxeter.NewIdentity(3)
	tab := tableau.New(id)
	result := svg.RenderTableau(tab)

	rectCount := strings.Count(result, "<rect")
	assert.Equal(t, 3, rectCount)

	for _, line := range strings.Split(result, "\n") {
		if strings.Contains(line, "<rect") {
			assert.True(t, strings.Contains(line, `width="80"`))
			assert.True(t, strings.Contains(line, `height="40"`))
		}
	}
}

func TestRenderTableau_Rank1(t *testing.T) {
	elem, _ := coxeter.NewElement([]int{1})
	tab := tableau.New(elem)
	result := svg.RenderTableau(tab)

	assert.True(t, strings.Contains(result, "<svg"))
	assert.Equal(t, 1, strings.Count(result, "<rect"))
}
