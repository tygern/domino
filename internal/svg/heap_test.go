package svg_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tygern/domino/internal/coxeter"
	"github.com/tygern/domino/internal/svg"
	"github.com/tygern/domino/internal/tableau"
)

func TestRenderHeap(t *testing.T) {
	elem, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	h := tableau.NewHeap(elem)
	result := svg.RenderHeap(h)

	assert.True(t, strings.Contains(result, "<svg"))
	assert.True(t, strings.Contains(result, "</svg>"))

	rectCount := strings.Count(result, "<rect")
	assert.Equal(t, elem.Length(), rectCount)
}

func TestRenderHeap_Dimensions(t *testing.T) {
	elem, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	h := tableau.NewHeap(elem)
	result := svg.RenderHeap(h)

	expectedWidth := h.MaxWidth() * 40 * 2
	expectedHeight := h.MaxHeight() * 40

	assert.True(t, strings.Contains(result, fmt.Sprintf(`width="%d"`, expectedWidth)))
	assert.True(t, strings.Contains(result, fmt.Sprintf(`height="%d"`, expectedHeight)))
}

func TestRenderHeap_PhantomSpacing(t *testing.T) {
	elem, _ := coxeter.NewElement([]int{-1, -2, 3, 4})
	h := tableau.NewHeap(elem)
	result := svg.RenderHeap(h)

	hasOffset1 := false
	hasOffset2 := false
	for _, line := range strings.Split(result, "\n") {
		if strings.Contains(line, ">1</text>") {
			hasOffset1 = true
		}
		if strings.Contains(line, ">2</text>") {
			hasOffset2 = true
		}
	}
	assert.True(t, hasOffset1)
	assert.True(t, hasOffset2)
}

func TestRenderHeap_SingleGenerator(t *testing.T) {
	elem, _ := coxeter.NewElement([]int{1, 3, 2, 4})
	h := tableau.NewHeap(elem)
	result := svg.RenderHeap(h)

	assert.Equal(t, 1, strings.Count(result, "<rect"))
}
