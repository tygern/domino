package web

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tygern/domino/internal/coxeter"
)

func parsePermutation(s string) (coxeter.Element, error) {
	parts := strings.Split(s, ",")
	perm := make([]int, len(parts))
	for i, p := range parts {
		val, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return coxeter.Element{}, fmt.Errorf("invalid integer %q", p)
		}
		perm[i] = val
	}
	return coxeter.NewElement(perm)
}

func parseExpression(s string, rank int) (coxeter.Element, error) {
	parts := strings.Split(s, ",")
	gens := make([]int, len(parts))
	for i, p := range parts {
		val, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return coxeter.Element{}, fmt.Errorf("invalid integer %q", p)
		}
		gens[i] = val
	}
	expr, err := coxeter.NewExpression(gens, rank)
	if err != nil {
		return coxeter.Element{}, err
	}
	return expr.ToElement(), nil
}

func formatSet(items []int) string {
	if len(items) == 0 {
		return "{}"
	}
	parts := make([]string, len(items))
	for i, v := range items {
		parts[i] = strconv.Itoa(v)
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func formatPermForURL(elem coxeter.Element) string {
	s := elem.String()
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	s = strings.ReplaceAll(s, " ", "")
	return s
}
