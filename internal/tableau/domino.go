package tableau

import "fmt"

type Domino struct {
	Label      int
	Col        int
	Row        int
	IsVertical bool
}

func (d Domino) SecondCol() int {
	if d.IsVertical {
		return d.Col
	}
	return d.Col + 1
}

func (d Domino) SecondRow() int {
	if d.IsVertical {
		return d.Row + 1
	}
	return d.Row
}

func (d Domino) MoveTo(col, row int) Domino {
	return Domino{Label: d.Label, Col: col, Row: row, IsVertical: d.IsVertical}
}

func (d Domino) Flip() Domino {
	return Domino{Label: d.Label, Col: d.Col, Row: d.Row, IsVertical: !d.IsVertical}
}

func (d Domino) Equal(other Domino) bool {
	return d.Label == other.Label && d.Col == other.Col && d.Row == other.Row && d.IsVertical == other.IsVertical
}

func (d Domino) String() string {
	orientation := "horizontal"
	if d.IsVertical {
		orientation = "vertical"
	}
	return fmt.Sprintf("Domino{%d at (%d,%d) %s}", d.Label, d.Col, d.Row, orientation)
}
