package coxeter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tygern/domino/internal/coxeter"
)

func TestExpression_ToElement(t *testing.T) {
	expr, err := coxeter.NewExpression([]int{3}, 4)
	assert.NoError(t, err)
	expected, _ := coxeter.NewElement([]int{1, 3, 2, 4})
	assert.True(t, expr.ToElement().Equal(expected))

	empty, err := coxeter.NewExpression([]int{}, 4)
	assert.NoError(t, err)
	assert.True(t, empty.ToElement().Equal(coxeter.NewIdentity(4)))
}

func TestExpression_NewExpression_Invalid(t *testing.T) {
	_, err := coxeter.NewExpression([]int{5}, 4)
	assert.Error(t, err)

	_, err = coxeter.NewExpression([]int{0}, 4)
	assert.Error(t, err)
}

func TestExpression_IsReduced(t *testing.T) {
	w, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	re := w.ReducedExpression()
	assert.True(t, re.IsReduced())

	nonReduced, _ := coxeter.NewExpression([]int{3, 3}, 4)
	assert.False(t, nonReduced.IsReduced())
}

func TestElement_ReducedExpression(t *testing.T) {
	perms := [][]int{
		{1, 2, 3, 4},
		{1, -4, 3, -2},
		{1, 2, 3, 4, 5, 6},
		{1, 3, -4, -2},
		{1, 3, 2, 4},
	}
	for _, p := range perms {
		elem, _ := coxeter.NewElement(p)
		re := elem.ReducedExpression()
		assert.True(t, elem.Equal(re.ToElement()), "round-trip failed for %v", p)
		assert.Equal(t, elem.Length(), re.Length(), "length mismatch for %v", p)
	}
}

func TestExpression_String(t *testing.T) {
	expr, _ := coxeter.NewExpression([]int{1, 3, 4, 3}, 4)
	assert.Equal(t, "(1, 3, 4, 3)", expr.String())
}

func TestExpression_Generators(t *testing.T) {
	expr, _ := coxeter.NewExpression([]int{1, 3, 4}, 4)
	gens := expr.Generators()
	assert.Equal(t, []int{1, 3, 4}, gens)

	gens[0] = 99
	assert.Equal(t, []int{1, 3, 4}, expr.Generators())
}
