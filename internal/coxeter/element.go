package coxeter

import (
	"fmt"
	"strings"
)

type Element struct {
	perm []int
}

func NewElement(perm []int) (Element, error) {
	n := len(perm)
	if n == 0 {
		return Element{}, fmt.Errorf("invalid signed permutation: empty")
	}

	seen := make([]bool, n)
	negCount := 0

	for _, v := range perm {
		abs := v
		if v < 0 {
			abs = -v
			negCount++
		}
		if v == 0 || abs > n || seen[abs-1] {
			return Element{}, fmt.Errorf("invalid signed permutation")
		}
		seen[abs-1] = true
	}
	if negCount%2 != 0 {
		return Element{}, fmt.Errorf("type D requires even number of negative values")
	}

	p := make([]int, n)
	copy(p, perm)
	return Element{perm: p}, nil
}

func NewIdentity(rank int) Element {
	perm := make([]int, rank)
	for i := range perm {
		perm[i] = i + 1
	}
	return Element{perm: perm}
}

func (e Element) Rank() int {
	return len(e.perm)
}

func (e Element) MapsTo(i int) int {
	n := len(e.perm)
	if i == 0 || abs(i) > n {
		return 0
	}
	if i > 0 {
		return e.perm[i-1]
	}
	return -e.perm[-i-1]
}

func (e Element) Equal(other Element) bool {
	if len(e.perm) != len(other.perm) {
		return false
	}
	for i := range e.perm {
		if e.perm[i] != other.perm[i] {
			return false
		}
	}
	return true
}

func (e Element) String() string {
	parts := make([]string, len(e.perm))
	for i, v := range e.perm {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func (e Element) Inverse() Element {
	n := len(e.perm)
	inv := make([]int, n)
	for i := 0; i < n; i++ {
		v := e.perm[i]
		if v > 0 {
			inv[v-1] = i + 1
		} else {
			inv[-v-1] = -(i + 1)
		}
	}
	return Element{perm: inv}
}

func (e Element) rightMultiplyGenerator(s int) Element {
	p := make([]int, len(e.perm))
	copy(p, e.perm)
	if s == 1 {
		p[0], p[1] = -p[1], -p[0]
	} else {
		p[s-2], p[s-1] = p[s-1], p[s-2]
	}
	return Element{perm: p}
}

func (e Element) RightMultiply(other Element) Element {
	n := len(e.perm)
	result := make([]int, n)
	for i := 1; i <= n; i++ {
		result[i-1] = e.MapsTo(other.MapsTo(i))
	}
	return Element{perm: result}
}

func (e Element) LeftMultiply(other Element) Element {
	return other.RightMultiply(e)
}

func countInversions(perm []int, factor int) int {
	count := 0
	n := len(perm)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if factor*perm[i] > perm[j] {
				count++
			}
		}
	}
	return count
}

func (e Element) Length() int {
	return countInversions(e.perm, 1) + countInversions(e.perm, -1)
}

func (e Element) IsRightDescent(s int) bool {
	if s == 1 {
		return -e.perm[1] > e.perm[0]
	}
	if s >= 2 && s <= len(e.perm) {
		return e.perm[s-2] > e.perm[s-1]
	}
	return false
}

func (e Element) RightDescentSet() []int {
	var descents []int
	for s := 1; s <= e.Rank(); s++ {
		if e.IsRightDescent(s) {
			descents = append(descents, s)
		}
	}
	return descents
}

func (e Element) LeftDescentSet() []int {
	return e.Inverse().RightDescentSet()
}

func (e Element) isRightBad() bool {
	n := len(e.perm)
	if n >= 3 && -e.perm[0] > e.perm[2] {
		return false
	}
	for j := 0; j <= n-3; j++ {
		if e.perm[j] > e.perm[j+2] {
			return false
		}
	}
	return true
}

func (e Element) isCommutingProduct() bool {
	n := len(e.perm)
	j := 0

	switch {
	case e.perm[0] == 1:
		j = 1
	case e.perm[0] == -1:
		if n < 2 || e.perm[1] != -2 {
			return false
		}
		j = 2
	case e.perm[0] == 2:
		if n < 2 || e.perm[1] != 1 {
			return false
		}
		j = 2
	case e.perm[0] == -2:
		if n < 2 || e.perm[1] != -1 {
			return false
		}
		j = 2
	default:
		return false
	}

	for j < n-2 {
		if e.perm[j] > j+2 {
			return false
		} else if e.perm[j] == j+1 {
			j++
		} else {
			if e.perm[j+1] != j+1 {
				return false
			}
			j += 2
		}
	}

	return true
}

func (e Element) IsBad() bool {
	if e.isCommutingProduct() {
		return false
	}
	return e.isRightBad() && e.Inverse().isRightBad()
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
