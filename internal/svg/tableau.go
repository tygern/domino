package svg

import (
	"fmt"
	"strings"

	"github.com/tygern/domino/internal/tableau"
)

const cellSize = 40

func RenderTableau(t tableau.Tableau) string {
	var b strings.Builder

	width := (t.MaxWidth() + 1) * cellSize
	height := (t.MaxHeight() + 1) * cellSize

	fmt.Fprintf(&b, `<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height)
	b.WriteString("\n")

	for _, d := range t.Dominoes() {
		x := (d.Col - 1) * cellSize
		y := (d.Row - 1) * cellSize

		var w, h int
		if d.IsVertical {
			w = cellSize
			h = cellSize * 2
		} else {
			w = cellSize * 2
			h = cellSize
		}

		cx := x + w/2
		cy := y + h/2

		fmt.Fprintf(&b, `  <rect x="%d" y="%d" width="%d" height="%d" fill="white" stroke="black" stroke-width="1.5"/>`, x, y, w, h)
		b.WriteString("\n")
		fmt.Fprintf(&b, `  <text x="%d" y="%d" text-anchor="middle" dominant-baseline="central" font-family="serif" font-size="16">%d</text>`, cx, cy, d.Label)
		b.WriteString("\n")
	}

	b.WriteString("</svg>")
	return b.String()
}
