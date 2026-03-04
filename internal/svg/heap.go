package svg

import (
	"fmt"
	"strings"

	"github.com/tygern/domino/internal/tableau"
)

func RenderHeap(h tableau.Heap) string {
	var b strings.Builder

	width := h.MaxWidth() * cellSize * 2
	height := h.MaxHeight() * cellSize

	fmt.Fprintf(&b, `<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height)
	b.WriteString("\n")

	for _, block := range h.Blocks() {
		x := block.Col * cellSize * 2
		y := (h.MaxHeight() - 1 - block.Row) * cellSize

		w := cellSize * 2
		ht := cellSize

		fmt.Fprintf(&b, `  <rect x="%d" y="%d" width="%d" height="%d" fill="white" stroke="black" stroke-width="1.5"/>`, x, y, w, ht)
		b.WriteString("\n")

		cx := x + w/2
		cy := y + ht/2

		switch block.Label {
		case 1:
			fmt.Fprintf(&b, `  <text x="%d" y="%d" text-anchor="middle" dominant-baseline="central" font-family="serif" font-size="16">1</text>`, cx-8, cy)
		case 2:
			fmt.Fprintf(&b, `  <text x="%d" y="%d" text-anchor="middle" dominant-baseline="central" font-family="serif" font-size="16">2</text>`, cx+8, cy)
		default:
			fmt.Fprintf(&b, `  <text x="%d" y="%d" text-anchor="middle" dominant-baseline="central" font-family="serif" font-size="16">%d</text>`, cx, cy, block.Label)
		}
		b.WriteString("\n")
	}

	b.WriteString("</svg>")
	return b.String()
}
