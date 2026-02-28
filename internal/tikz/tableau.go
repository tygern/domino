package tikz

import (
	"fmt"
	"strings"

	"github.com/tygern/domino/internal/tableau"
)

func RenderTableau(t tableau.Tableau) string {
	var b strings.Builder
	b.WriteString("\\begin{tikzpicture}[node distance=0 cm,outer sep = 0pt]\n")
	b.WriteString("\\tikzstyle{ver}=[rectangle, draw, thick, minimum width=1cm, minimum height=2cm]\n")
	b.WriteString("\\tikzstyle{hor}=[rectangle, draw, thick, minimum width=2cm, minimum height=1cm]\n")

	for _, d := range t.Dominoes() {
		if d.IsVertical {
			fmt.Fprintf(&b, "\\node[ver] at (0 + %d, 4 - %d) {%d};\n", d.Col, d.Row, d.Label)
		} else {
			fmt.Fprintf(&b, "\\node[hor] at (.5 + %d, 4.5 - %d) {%d};\n", d.Col, d.Row, d.Label)
		}
	}

	b.WriteString("\\end{tikzpicture}\n")
	return b.String()
}
