package main

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func domino(args ...string) (string, string, error) {
	cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestMain_Info(t *testing.T) {
	stdout, _, err := domino("info", "-perm", "1,-4,3,-2")
	assert.NoError(t, err)
	assert.Contains(t, stdout, "Length:        7")
	assert.Contains(t, stdout, "Right descent: {1, 2, 4}")
	assert.Contains(t, stdout, "Bad:           true")
	assert.Contains(t, stdout, "Reduced:")
}

func TestMain_Info_Expr(t *testing.T) {
	stdout, _, err := domino("info", "-expr", "3", "-rank", "4")
	assert.NoError(t, err)
	assert.Contains(t, stdout, "[1, 3, 2, 4]")
	assert.Contains(t, stdout, "Length:        1")
}

func TestMain_Tableau(t *testing.T) {
	stdout, _, err := domino("tableau", "-perm", "1,-4,3,-2")
	assert.NoError(t, err)
	assert.Contains(t, stdout, "Right tableau:")
	assert.Contains(t, stdout, "Left tableau:")
	assert.Contains(t, stdout, "\\begin{tikzpicture}")
}

func TestMain_Heap(t *testing.T) {
	stdout, _, err := domino("heap", "-perm", "1,-4,3,-2")
	assert.NoError(t, err)
	assert.Contains(t, stdout, "\\begin{tikzpicture}")
	assert.Contains(t, stdout, "\\end{tikzpicture}")
}

func TestMain_Bad(t *testing.T) {
	stdout, stderr, err := domino("bad", "-rank", "4")
	assert.NoError(t, err)
	assert.Contains(t, stdout, "[1, -4, 3, -2]")
	assert.Contains(t, stderr, "1 bad elements in D_4")
}

func TestMain_NoArgs(t *testing.T) {
	_, stderr, err := domino()
	assert.Error(t, err)
	assert.Contains(t, stderr, "Usage:")
}

func TestMain_InvalidCommand(t *testing.T) {
	_, stderr, err := domino("invalid")
	assert.Error(t, err)
	assert.Contains(t, stderr, "Usage:")
}

func TestMain_Info_MissingPerm(t *testing.T) {
	_, stderr, err := domino("info")
	assert.Error(t, err)
	assert.Contains(t, stderr, "Error:")
}

func TestMain_Bad_MissingRank(t *testing.T) {
	_, stderr, err := domino("bad")
	assert.Error(t, err)
	assert.Contains(t, stderr, "-rank is required")
}

func TestMain_Info_InvalidPerm(t *testing.T) {
	_, stderr, err := domino("info", "-perm", "1,abc,3")
	assert.Error(t, err)
	assert.Contains(t, stderr, "Error:")
}

func TestMain_Bad_InvalidRank(t *testing.T) {
	_, stderr, err := domino("bad", "-rank", "0")
	assert.Error(t, err)
	assert.Contains(t, stderr, "invalid rank")
}
