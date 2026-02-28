package coxeter

import (
	"fmt"
	"slices"
	"strings"
)

type Expression struct {
	generators []int
	rank       int
}

func NewExpression(generators []int, rank int) (Expression, error) {
	for _, g := range generators {
		if g < 1 || g > rank {
			return Expression{}, fmt.Errorf("generator %d out of range [1, %d]", g, rank)
		}
	}
	gens := make([]int, len(generators))
	copy(gens, generators)
	return Expression{generators: gens, rank: rank}, nil
}

func (ex Expression) Generators() []int {
	cp := make([]int, len(ex.generators))
	copy(cp, ex.generators)
	return cp
}

func (ex Expression) Rank() int {
	return ex.rank
}

func (ex Expression) Length() int {
	return len(ex.generators)
}

func (ex Expression) String() string {
	parts := make([]string, len(ex.generators))
	for i, g := range ex.generators {
		parts[i] = fmt.Sprintf("%d", g)
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func (ex Expression) ToElement() Element {
	result := NewIdentity(ex.rank)
	for _, s := range ex.generators {
		result = result.rightMultiplyGenerator(s)
	}
	return result
}

func (ex Expression) IsReduced() bool {
	return ex.Length() == ex.ToElement().Length()
}

func (e Element) ReducedExpression() Expression {
	var generators []int
	current := e
	for current.Length() > 0 {
		for s := current.Rank(); s >= 1; s-- {
			if current.IsRightDescent(s) {
				generators = append(generators, s)
				current = current.rightMultiplyGenerator(s)
			}
		}
	}
	slices.Reverse(generators)
	return Expression{generators: generators, rank: e.Rank()}
}
