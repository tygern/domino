package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tygern/domino/internal/coxeter"
)

func TestParsePermutation(t *testing.T) {
	elem, err := parsePermutation("1, -4, 3, -2")
	assert.NoError(t, err)

	expected, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	assert.True(t, elem.Equal(expected))
}

func TestParsePermutation_InvalidOddNegations(t *testing.T) {
	_, err := parsePermutation("1, -2, 3")
	assert.Error(t, err)
}

func TestParsePermutation_NonInteger(t *testing.T) {
	_, err := parsePermutation("a, b")
	assert.Error(t, err)
}

func TestParseExpression(t *testing.T) {
	elem, err := parseExpression("3", 4)
	assert.NoError(t, err)

	expected, _ := coxeter.NewElement([]int{1, 3, 2, 4})
	assert.True(t, elem.Equal(expected))
}

func TestParseExpression_OutOfRange(t *testing.T) {
	_, err := parseExpression("5", 4)
	assert.Error(t, err)
}

func TestFormatSet(t *testing.T) {
	assert.Equal(t, "{}", formatSet(nil))
	assert.Equal(t, "{1, 2, 4}", formatSet([]int{1, 2, 4}))
}

func TestFormatPermForURL(t *testing.T) {
	elem, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	assert.Equal(t, "1,-4,3,-2", formatPermForURL(elem))
}
