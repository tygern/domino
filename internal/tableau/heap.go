package tableau

import (
	"github.com/tygern/domino/internal/coxeter"
)

type Heap struct {
	blocks []Domino
	rank   int
	width  int
	height int
}

func NewHeap(elem coxeter.Element) Heap {
	expr := elem.ReducedExpression()
	length := expr.Length()

	blocks := make([]Domino, length)
	heights := make([]int, expr.Rank()+1)
	firstBlockUsed := make([]bool, length+2)
	width, height := 0, 0

	gens := expr.Generators()
	for i, gen := range gens {
		var col, row int

		if gen <= 2 {
			col = 0
			if firstBlockUsed[heights[1]] {
				row = heights[1] - 1
			} else {
				row = heights[1]
				heights[1]++
				firstBlockUsed[heights[1]] = true
			}
		} else {
			col = gen - 2
			row = heights[col]
			if heights[col+1] > row {
				row = heights[col+1]
			}
			heights[col] = row + 1
			heights[col+1] = row + 1
		}

		if row+1 > height {
			height = row + 1
		}
		if gen > width {
			width = gen
		}

		blocks[i] = Domino{Label: gen, Col: col, Row: row}
	}

	if width == 1 {
		width = 2
	}

	return Heap{blocks: blocks, rank: expr.Rank(), width: width, height: height}
}

func (h Heap) Blocks() []Domino {
	cp := make([]Domino, len(h.blocks))
	copy(cp, h.blocks)
	return cp
}

func (h Heap) Rank() int {
	return h.rank
}

func (h Heap) Size() int {
	return len(h.blocks)
}

func (h Heap) MaxWidth() int {
	return h.width
}

func (h Heap) MaxHeight() int {
	return h.height
}
