package tableau

import (
	"github.com/tygern/domino/internal/coxeter"
)

type Tableau struct {
	rank     int
	dominoes []Domino
	present  []bool
	grid     [][]int
	maxLabel int
}

func New(elem coxeter.Element) Tableau {
	rank := elem.Rank()
	gridSize := 2*rank + 2
	t := Tableau{
		rank:     rank,
		dominoes: make([]Domino, rank),
		present:  make([]bool, rank),
		grid:     makeGrid(gridSize),
	}

	for i := 1; i <= rank; i++ {
		val := elem.MapsTo(i)
		isVertical := val < 0
		label := val
		if label < 0 {
			label = -label
		}
		d := Domino{Label: label, IsVertical: isVertical}
		t = t.addDomino(d)
	}
	return t
}

func RightTableau(elem coxeter.Element) Tableau {
	return New(elem)
}

func LeftTableau(elem coxeter.Element) Tableau {
	return New(elem.Inverse())
}

func (t Tableau) Rank() int {
	return t.rank
}

func (t Tableau) Size() int {
	count := 0
	for _, p := range t.present {
		if p {
			count++
		}
	}
	return count
}

func (t Tableau) MaxWidth() int {
	return t.largestInRow(1, t.rank+1)
}

func (t Tableau) MaxHeight() int {
	return t.largestInCol(1, t.rank+1)
}

func (t Tableau) GetDomino(label int) (Domino, bool) {
	if label < 1 || label > t.rank || !t.present[label-1] {
		return Domino{}, false
	}
	return t.dominoes[label-1], true
}

func (t Tableau) Dominoes() []Domino {
	var result []Domino
	for i := 0; i < t.rank; i++ {
		if t.present[i] {
			result = append(result, t.dominoes[i])
		}
	}
	return result
}

func (t Tableau) Equal(other Tableau) bool {
	if t.rank != other.rank {
		return false
	}
	for i := 0; i < t.rank; i++ {
		if t.present[i] != other.present[i] {
			return false
		}
		if t.present[i] && !t.dominoes[i].Equal(other.dominoes[i]) {
			return false
		}
	}
	return true
}

func makeGrid(size int) [][]int {
	grid := make([][]int, size)
	for i := range grid {
		grid[i] = make([]int, size)
	}
	return grid
}

func (t Tableau) placeDomino(d Domino) Tableau {
	t.dominoes[d.Label-1] = d
	t.present[d.Label-1] = true
	t.setGrid(d.Col, d.Row, d.Label)
	t.setGrid(d.SecondCol(), d.SecondRow(), d.Label)
	if d.Label > t.maxLabel {
		t.maxLabel = d.Label
	}
	return t
}

func (t Tableau) removeDomino(label int) Tableau {
	d := t.dominoes[label-1]
	if t.gridAt(d.Col, d.Row) == label {
		t.setGrid(d.Col, d.Row, 0)
	}
	if t.gridAt(d.SecondCol(), d.SecondRow()) == label {
		t.setGrid(d.SecondCol(), d.SecondRow(), 0)
	}
	t.present[label-1] = false
	return t
}

func (t Tableau) setGrid(col, row, val int) {
	if col >= len(t.grid) || row >= len(t.grid[0]) {
		t.growGrid(col, row)
	}
	t.grid[col][row] = val
}

func (t *Tableau) growGrid(needCol, needRow int) {
	newSize := len(t.grid)
	for newSize <= needCol || newSize <= needRow {
		newSize *= 2
	}

	newGrid := make([][]int, newSize)
	for i := range newGrid {
		newGrid[i] = make([]int, newSize)
	}
	for i := range t.grid {
		copy(newGrid[i], t.grid[i])
	}
	t.grid = newGrid
}

func (t Tableau) gridAt(col, row int) int {
	if col >= len(t.grid) || row >= len(t.grid[0]) || col < 0 || row < 0 {
		return 0
	}
	return t.grid[col][row]
}

func (t Tableau) largestInRow(row, bound int) int {
	maxCol := 0
	for col := 1; col < len(t.grid); col++ {
		label := t.grid[col][row]
		if label > 0 && label < bound {
			if col > maxCol {
				maxCol = col
			}
		}
	}
	return maxCol
}

func (t Tableau) largestInCol(col, bound int) int {
	maxRow := 0
	if col < len(t.grid) {
		for row := 1; row < len(t.grid[col]); row++ {
			label := t.grid[col][row]
			if label > 0 && label < bound {
				if row > maxRow {
					maxRow = row
				}
			}
		}
	}
	return maxRow
}

func (t Tableau) overlapCount(label int) int {
	d := t.dominoes[label-1]
	count := 0
	g1 := t.gridAt(d.Col, d.Row)
	if g1 > 0 && g1 != label {
		count++
	}
	g2 := t.gridAt(d.SecondCol(), d.SecondRow())
	if g2 > 0 && g2 != label {
		count++
	}
	return count
}

func (t Tableau) addDomino(d Domino) Tableau {
	label := d.Label

	if label > t.maxLabel {
		if d.IsVertical {
			d = d.MoveTo(1, t.largestInCol(1, label)+1)
		} else {
			d = d.MoveTo(t.largestInRow(1, label)+1, 1)
		}
		t = t.placeDomino(d)
	} else {
		if d.IsVertical {
			d = d.MoveTo(1, t.largestInCol(1, label)+1)
		} else {
			d = d.MoveTo(t.largestInRow(1, label)+1, 1)
		}
		t = t.placeDomino(d)

		for j := label + 1; j <= t.rank; j++ {
			if t.present[j-1] {
				t = t.shuffle(j)
			}
		}
	}
	return t
}

func (t Tableau) shuffle(label int) Tableau {
	d := t.dominoes[label-1]
	row := d.Row
	col := d.Col
	overlap := t.overlapCount(label)

	if overlap == 0 {
		return t
	}

	t = t.removeDomino(label)

	if overlap == 1 {
		if d.IsVertical {
			newRow := row + 1
			newCol := t.largestInRow(newRow, label) + 1
			d = Domino{Label: label, Col: newCol, Row: newRow, IsVertical: false}
		} else {
			newCol := col + 1
			newRow := t.largestInCol(newCol, label) + 1
			d = Domino{Label: label, Col: newCol, Row: newRow, IsVertical: true}
		}
	} else {
		if d.IsVertical {
			newCol := col + 1
			newRow := t.largestInCol(newCol, label) + 1
			d = Domino{Label: label, Col: newCol, Row: newRow, IsVertical: true}
		} else {
			newRow := row + 1
			newCol := t.largestInRow(newRow, label) + 1
			d = Domino{Label: label, Col: newCol, Row: newRow, IsVertical: false}
		}
	}

	return t.placeDomino(d)
}
