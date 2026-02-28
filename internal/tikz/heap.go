package tikz

import (
	"fmt"
	"strings"

	"github.com/tygern/domino/internal/tableau"
)

func RenderHeap(h tableau.Heap) string {
	var b strings.Builder
	b.WriteString("\\begin{tikzpicture}[node distance=0 cm,outer sep = 0pt]\n")
	b.WriteString("\\tikzstyle{hor}=[rectangle, draw, thick, minimum width=2cm, minimum height=1cm]\n")

	for _, block := range h.Blocks() {
		x := block.Col
		y := block.Row

		switch block.Label {
		case 1:
			fmt.Fprintf(&b, "\\node[hor] at ( %d, %d) {1\\phantom{ 2}};\n", x, y)
		case 2:
			fmt.Fprintf(&b, "\\node[hor] at ( %d, %d) {\\phantom{1 }2};\n", x, y)
		default:
			fmt.Fprintf(&b, "\\node[hor] at ( %d, %d) {%d};\n", x, y, block.Label)
		}
	}

	b.WriteString("\\end{tikzpicture}\n")
	return b.String()
}
