# Plan: BuildDomino Go Rewrite

## Package structure

```
domino/
├── cmd/
│   └── domino/
│       └── main.go              # CLI entry point
├── internal/
│   ├── coxeter/
│   │   ├── element.go           # TypeD signed permutation
│   │   ├── element_test.go
│   │   ├── expression.go        # Reduced expressions (generator sequences)
│   │   └── expression_test.go
│   ├── tableau/
│   │   ├── domino.go            # Domino and grid value types
│   │   ├── tableau.go           # Garfinkle tableau construction
│   │   ├── tableau_test.go
│   │   ├── heap.go              # Heap construction
│   │   └── heap_test.go
│   └── tikz/
│       ├── tableau.go           # TikZ rendering for tableaux
│       ├── heap.go              # TikZ rendering for heaps
│       └── tikz_test.go
├── go.mod
├── CLAUDE.md
├── research.md
└── plan.md
```

## Phase 1: Core types — `internal/coxeter`

### element.go

A type D element is a signed permutation stored as a slice of ints. The slice
represents one-line notation: `perm[i]` is the image of `i+1`, and negative
values encode sign changes. The constraint for type D is that the count of
negative values must be even.

```go
package coxeter

type Element struct {
	perm []int // one-line notation, 0-indexed: perm[i] = image of (i+1)
}
```

The constructor validates the input: no zeros, no out-of-range values, no
repeated absolute values, and an even number of negatives.

```go
func NewElement(perm []int) (Element, error) {
	n := len(perm)
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
```

Identity constructor:

```go
func NewIdentity(rank int) Element {
	perm := make([]int, rank)
	for i := range perm {
		perm[i] = i + 1
	}
	return Element{perm: perm}
}
```

Methods on Element — all return new values, no mutation:

| Method | Signature | Description |
|--------|-----------|-------------|
| Rank | `Rank() int` | `len(e.perm)` |
| MapsTo | `MapsTo(i int) int` | Image of i (supports negative i) |
| Inverse | `Inverse() Element` | Inverts the permutation |
| RightMultiply | `RightMultiply(other Element) Element` | Composition: `e(other(i))` |
| LeftMultiply | `LeftMultiply(other Element) Element` | `other.RightMultiply(e)` |
| Length | `Length() int` | Coxeter length = inversions(+1) + inversions(-1) |
| IsRightDescent | `IsRightDescent(s int) bool` | Tests generator s |
| RightDescentSet | `RightDescentSet() []int` | All right descents |
| LeftDescentSet | `LeftDescentSet() []int` | `e.Inverse().RightDescentSet()` |
| IsBad | `IsBad() bool` | Bad element detection |
| String | `String() string` | One-line notation: `[1, -4, 3, -2]` |
| Equal | `Equal(other Element) bool` | Slice comparison |

#### Generator multiplication

Multiplying by generator s on the right:
- s=1: swap positions 0 and 1, negate both
- s>=2: swap positions s-2 and s-1

This is an internal helper. It returns a new Element (no mutation).

```go
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
```

#### Right descent test

```go
func (e Element) IsRightDescent(s int) bool {
	if s == 1 {
		return -e.perm[1] > e.perm[0]
	}
	return e.perm[s-2] > e.perm[s-1]
}
```

#### Coxeter length

The Java code computes `countInv(1) + countInv(-1)`. The `countInv(factor)`
function counts pairs (i,j) with i<j where `factor*perm[i] > perm[j]`. The Java
implementation uses a divide-and-conquer approach but with an O(n^2) merge step,
so it's actually O(n^2) despite looking like mergesort. For the rewrite, we can
use proper merge-sort inversion counting for O(n log n), or keep O(n^2) since
ranks are small.

For simplicity and correctness, start with the direct O(n^2) count:

```go
func countInversions(perm []int, factor int) int {
	count := 0
	for i := 0; i < len(perm); i++ {
		for j := i + 1; j < len(perm); j++ {
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
```

#### Bad element detection

A bad element is one that:
1. Is NOT a product of commuting generators, AND
2. Has no reduced expression starting in two noncommuting generators (left bad), AND
3. Has no reduced expression ending in two noncommuting generators (right bad)

`isRightBad` checks whether every reduced expression ends in commuting generators.
From the Java code, the logic is:
- If `-perm[0] > perm[2]`, NOT right-bad (can end in s1,s3 or s3,s1)
- For each j from 0 to n-3: if `perm[j] > perm[j+2]`, NOT right-bad
- Otherwise, right-bad

`isLeftBad` = inverse is right-bad.

`commutingGenerators` checks if the element is a product of pairwise commuting
generators. The Java code checks a specific structural pattern in the one-line
notation: the first two positions can be identity, swap (positions 1,2), negate
(positions 1,2), or negate-swap; then remaining positions must be either fixed
points or adjacent transpositions with no overlapping.

```go
func (e Element) IsBad() bool {
	if e.isCommutingProduct() {
		return false
	}
	return e.isRightBad() && e.Inverse().isRightBad()
}
```

#### Reduced expression generation

The `findRE` algorithm builds a reduced expression by repeatedly stripping right
descents from highest to lowest generator index:

```go
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
	// Reverse to get left-to-right order
	slices.Reverse(generators)
	return Expression{generators: generators, rank: e.Rank()}
}
```

#### Enumerate all bad elements

New feature. To enumerate all elements of D_n, we generate all signed
permutations of {1,...,n} with an even number of negatives, then filter by
`IsBad()`.

A signed permutation of n elements is a permutation of {1,...,n} combined with a
sign assignment to each position. For type D, the number of negative signs must
be even.

Total elements in D_n = `n! * 2^(n-1)`. This grows fast:
- D_3: 24
- D_4: 192
- D_5: 1920
- D_6: 23040

Strategy: generate all permutations of {1,...,n}, for each permutation generate
all even-sized subsets of positions to negate, check `IsBad()`.

The naive approach generates all `n! * 2^(n-1)` elements and filters. This is
fine for small ranks but unusable beyond ~8:

| Rank | Elements | Time (naive) |
|------|----------|-------------|
| 4 | 192 | instant |
| 6 | 23,040 | instant |
| 8 | 2,580,480 | seconds |
| 10 | 368,640,000 | minutes |
| 12 | 6.7 * 10^10 | impossible |

We need to prune the search space. The key observation is that `isRightBad` and
`isCommutingProduct` can both be checked on partial permutations, and most
elements fail the bad test early.

#### Optimization 1: Prune on `isRightBad` during permutation generation

Recall `isRightBad` returns false (i.e., NOT right-bad, so NOT bad) if:
- `-perm[0] > perm[2]`, or
- `perm[j] > perm[j+2]` for any j in 0..n-3

This means as soon as we've placed values at positions 0 and 2, we can check
whether `-perm[0] > perm[2]`. If true, the element cannot be right-bad, and
therefore cannot be bad regardless of what we put in the remaining positions. We
can skip the entire subtree.

More generally, as soon as positions j and j+2 are both filled and
`perm[j] > perm[j+2]`, we can skip.

Build the signed permutation position by position using backtracking. At each
step, check the partial right-bad conditions for any newly completable pairs:

```go
func BadElements(rank int) []Element {
	var result []Element
	perm := make([]int, rank)
	used := make([]bool, rank+1) // used[v] = true if |v| already placed
	badElements(perm, used, 0, rank, &result)
	return result
}

func badElements(perm []int, used []bool, pos, rank int, result *[]Element) {
	if pos == rank {
		elem, err := NewElement(perm)
		if err != nil {
			return
		}
		if elem.IsBad() {
			cp := make([]int, rank)
			copy(cp, perm)
			*result = append(*result, Element{perm: cp})
		}
		return
	}

	for absVal := 1; absVal <= rank; absVal++ {
		if used[absVal] {
			continue
		}
		used[absVal] = true

		for _, sign := range []int{1, -1} {
			perm[pos] = sign * absVal

			if canBeRightBad(perm, pos, rank) {
				badElements(perm, used, pos+1, rank, result)
			}
		}

		used[absVal] = false
	}
}
```

The pruning function checks partial right-bad conditions:

```go
func canBeRightBad(perm []int, filledUpTo, rank int) bool {
	// Check the special condition: -perm[0] > perm[2]
	if filledUpTo >= 2 {
		if -perm[0] > perm[2] {
			return false
		}
	}

	// Check perm[j] > perm[j+2] for all j where both are filled
	for j := 0; j+2 <= filledUpTo; j++ {
		if perm[j] > perm[j+2] {
			return false
		}
	}

	return true
}
```

This prunes entire subtrees as soon as any "not right-bad" condition is
triggered. In practice, the `-perm[0] > perm[2]` check alone eliminates roughly
half the search space at depth 3.

#### Optimization 2: Prune on even-negatives constraint

Rather than generating all sign patterns and filtering, track the negative count
during generation. At each position, if the remaining positions can't make the
total negative count even, skip:

```go
func canAchieveEvenNeg(negCount, pos, rank int) bool {
	remaining := rank - pos - 1
	// Need (negCount + future negatives) to be even
	// If remaining == 0, negCount must already be even
	// Otherwise, we can always adjust (at least 0 or 1 more negatives)
	if remaining == 0 {
		return negCount%2 == 0
	}
	return true // can always fix parity with remaining positions
}
```

Actually this only constrains the last position: if we reach `pos == rank-1`
with an odd negative count, we must negate the last value; if even, we must not.
This cuts the last branch factor from `2*remaining_values` to `remaining_values`.

#### Optimization 3: Prune on left-bad too

A bad element must be BOTH right-bad and left-bad. Left-bad means the inverse is
right-bad. While we can't fully check the inverse until the permutation is
complete, we CAN check partial inverse conditions as positions are filled.

When we set `perm[pos] = v`, the inverse has `inv[|v|-1] = sign(v) * (pos+1)`.
So filling position `pos` fills position `|v|-1` of the inverse. We can apply
the same `canBeRightBad` check to the partial inverse.

```go
func badElements(perm, inv []int, used []bool, pos, rank, negCount int, result *[]Element) {
	if pos == rank {
		// Both perm and inv are complete — check full IsBad
		elem := Element{perm: slices.Clone(perm)}
		if elem.IsBad() {
			*result = append(*result, elem)
		}
		return
	}

	for absVal := 1; absVal <= rank; absVal++ {
		if used[absVal] {
			continue
		}
		used[absVal] = true

		for _, sign := range []int{1, -1} {
			newNeg := negCount
			if sign < 0 {
				newNeg++
			}

			// Last position: enforce even parity
			if pos == rank-1 && newNeg%2 != 0 {
				continue
			}

			perm[pos] = sign * absVal

			// Set inverse: inv[absVal-1] = sign * (pos+1)
			inv[absVal-1] = sign * (pos + 1)

			if canBeRightBad(perm, pos, rank) && canBeRightBad(inv, absVal-1, rank) {
				badElements(perm, inv, used, pos+1, rank, newNeg, result)
			}

			inv[absVal-1] = 0
		}

		used[absVal] = false
	}
}
```

Note: `canBeRightBad` for the inverse needs a slight adjustment — it should only
check pairs where BOTH positions are filled (non-zero). Since the inverse is
filled out of order, we track which positions have been set:

```go
func canBeRightBadPartial(perm []int, rank int) bool {
	if perm[0] != 0 && perm[2] != 0 {
		if -perm[0] > perm[2] {
			return false
		}
	}
	for j := 0; j+2 < rank; j++ {
		if perm[j] != 0 && perm[j+2] != 0 && perm[j] > perm[j+2] {
			return false
		}
	}
	return true
}
```

#### Optimization 4: Parallelism

The outermost loop (choice of `perm[0]`) can be parallelized across goroutines.
Each goroutine explores an independent subtree. For rank n, there are `2n`
choices for position 0 (values ±1 through ±n), giving natural parallelism:

```go
func BadElements(rank int) []Element {
	var mu sync.Mutex
	var result []Element
	var wg sync.WaitGroup

	for absVal := 1; absVal <= rank; absVal++ {
		for _, sign := range []int{1, -1} {
			wg.Add(1)
			go func(startVal int) {
				defer wg.Done()
				local := badElementsFrom(startVal, rank)
				mu.Lock()
				result = append(result, local...)
				mu.Unlock()
			}(sign * absVal)
		}
	}
	wg.Wait()
	return result
}
```

#### Expected speedup

The combination of these optimizations should make rank 10-12 feasible:
- Right-bad pruning eliminates most of the search tree at shallow depth
- Inverse pruning (left-bad) adds a second independent cut
- Even-parity constraint halves the last branch
- Parallelism gives ~2n× throughput on multi-core machines

For very large ranks (15+), the search space is still astronomical and would
require mathematical analysis to further reduce candidates — but that's beyond
the scope of this tool.

### expression.go

A reduced expression is a sequence of generator indices.

```go
type Expression struct {
	generators []int
	rank       int
}
```

Methods:

| Method | Signature | Description |
|--------|-----------|-------------|
| Generators | `Generators() []int` | Returns a copy of the generator sequence |
| Rank | `Rank() int` | Rank of the Coxeter group |
| Length | `Length() int` | Number of generators |
| ToElement | `ToElement() Element` | Convert to signed permutation |
| IsReduced | `IsReduced() bool` | Check if expression is reduced |
| String | `String() string` | Format as `(1, 3, 4, 3)` |

`ToElement` applies generators left to right starting from the identity:

```go
func (ex Expression) ToElement() Element {
	result := NewIdentity(ex.rank)
	for _, s := range ex.generators {
		result = result.rightMultiplyGenerator(s)
	}
	return result
}
```

### element_test.go

Port every test case from `TypeDTest.java` and `TypeDExpressionTest.java`.
Test names follow `TestElement_MethodName` convention. Key test cases to port:

```go
func TestElement_NewElement(t *testing.T) {
	// Valid: identity
	_, err := coxeter.NewElement([]int{1, 2, 3, 4})
	assert.NoError(t, err)

	// Valid: even negatives
	_, err = coxeter.NewElement([]int{1, -4, 3, -2})
	assert.NoError(t, err)

	// Invalid: odd negatives
	_, err = coxeter.NewElement([]int{-4, 1, 2, 3})
	assert.Error(t, err)

	// Invalid: out of range
	_, err = coxeter.NewElement([]int{-7, 1, -2, 3})
	assert.Error(t, err)

	// Invalid: zero
	_, err = coxeter.NewElement([]int{-4, 0, -2, 3})
	assert.Error(t, err)

	// Invalid: repeated absolute value
	_, err = coxeter.NewElement([]int{2, 1, -2, 3})
	assert.Error(t, err)
}

func TestElement_IsBad(t *testing.T) {
	tests := []struct {
		perm []int
		bad  bool
	}{
		{[]int{1, -4, 3, -2}, true},
		{[]int{-1, -6, 3, -4, 5, -2}, true},
		{[]int{1, -8, 3, -6, 5, -4, 7, -2}, true},
		{[]int{1, 2, 3, 4}, false},
		{[]int{1, 3, -4, -2}, false},
		{[]int{1, 3, 2, 4}, false},
		{[]int{-1, 3, 4, -2}, false}, // commuting generators
	}
	for _, tt := range tests {
		elem, err := coxeter.NewElement(tt.perm)
		assert.NoError(t, err)
		assert.Equal(t, tt.bad, elem.IsBad(), "perm=%v", tt.perm)
	}
}

func TestElement_Length(t *testing.T) {
	tests := []struct {
		perm   []int
		length int
	}{
		{[]int{1, 2, 3, 4}, 0},
		{[]int{1, -4, 3, -2}, 7},
		{[]int{1, 3, 2, 4}, 1},
		{[]int{1, 3, -4, -2}, 8},
	}
	for _, tt := range tests {
		elem, _ := coxeter.NewElement(tt.perm)
		assert.Equal(t, tt.length, elem.Length())
	}
}

func TestElement_ReducedExpression(t *testing.T) {
	// A reduced expression, when converted back, gives the same element
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
		assert.True(t, elem.Equal(re.ToElement()))
		assert.Equal(t, elem.Length(), re.Length())
	}
}
```

Test for bad element enumeration:

```go
func TestBadElements(t *testing.T) {
	// D_3 has no bad elements (only D_4+ has them)
	assert.Empty(t, coxeter.BadElements(3))

	// D_4 should have a known count of bad elements
	bad4 := coxeter.BadElements(4)
	for _, elem := range bad4 {
		assert.True(t, elem.IsBad())
	}

	// Verify known bad element is in the list
	known, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	found := false
	for _, elem := range bad4 {
		if elem.Equal(known) {
			found = true
			break
		}
	}
	assert.True(t, found)
}
```

## Phase 2: Tableau construction — `internal/tableau`

### domino.go

A domino is an immutable value type: a label, a position (col, row of the first
block), and an orientation.

```go
package tableau

type Domino struct {
	Label      int
	Col        int // x-coordinate of first block
	Row        int // y-coordinate of first block
	IsVertical bool
}
```

The second block position is derived:

```go
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
```

No setters, no mutation. To "move" or "flip" a domino, create a new one:

```go
func (d Domino) MoveTo(col, row int) Domino {
	return Domino{Label: d.Label, Col: col, Row: row, IsVertical: d.IsVertical}
}

func (d Domino) Flip() Domino {
	return Domino{Label: d.Label, Col: d.Col, Row: d.Row, IsVertical: !d.IsVertical}
}
```

### tableau.go

The Tableau holds dominoes in a slice indexed by label, plus a 2D grid for O(1)
positional lookups. The grid maps (col, row) to the label of the domino
occupying that cell (0 means empty).

```go
type Tableau struct {
	rank     int
	dominoes []Domino  // indexed by label-1, zero value means absent
	present  []bool    // present[i] = true if domino i+1 is in tableau
	grid     [][]int   // grid[col][row] = label occupying that cell, 0 = empty
}
```

The grid eliminates the O(n) scans that `largestInRow` and `largestInCol` do
in the Java version. With a grid, finding the rightmost occupied cell in a row
is still a scan of that row, but the row length is bounded by the tableau width
(much smaller than n for most shapes). More importantly, overlap checks become
O(1) — just look up the two cells the domino would occupy.

Constructor from a coxeter.Element:

```go
func New(elem coxeter.Element) Tableau {
	rank := elem.Rank()
	t := Tableau{
		rank:     rank,
		dominoes: make([]Domino, rank),
		present:  make([]bool, rank),
		grid:     makeGrid(rank),
	}

	for i := 1; i <= rank; i++ {
		val := elem.MapsTo(i)
		isVertical := val < 0
		d := Domino{Label: abs(val), IsVertical: isVertical}
		t = t.addDomino(d)
	}
	return t
}
```

The `addDomino` method implements Garfinkle's algorithm. The key insight from the
Java code:

1. If the new domino has the largest label seen so far, place it at the end of
   row 1 (horizontal) or column 1 (vertical).
2. Otherwise, compute where it would go ignoring all larger-labeled dominoes
   (the alpha-map), place it there, then shuffle all larger-labeled dominoes.

The shuffle operation for a domino with label `j`:
- Count how many cells of domino `j` overlap with cells occupied by
  smaller-labeled dominoes.
- overlap=0: do nothing
- overlap=1: twist (flip orientation, move to next row/column)
- overlap=2: slide (keep orientation, move to next row/column)

```go
func (t Tableau) addDomino(d Domino) Tableau {
	label := d.Label

	if !t.hasLarger(label) {
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
```

The shuffle method:

```go
func (t Tableau) shuffle(label int) Tableau {
	d := t.dominoes[label-1]
	overlap := t.overlapCount(label)

	if overlap == 0 {
		return t
	}

	t = t.removeDomino(label)

	if overlap == 1 {
		// Twist: flip and move
		d = d.Flip()
		if !d.IsVertical {
			// Was vertical, now horizontal — move to next row
			col := t.largestInRow(d.Row, label) + 1
			d = d.MoveTo(col, d.Row)
		} else {
			// Was horizontal, now vertical — move to next column
			row := t.largestInCol(d.Col, label) + 1
			d = d.MoveTo(d.Col, row)
		}
	} else {
		// Slide: keep orientation, move
		if d.IsVertical {
			col := d.Col + 1
			row := t.largestInCol(col, label) + 1
			d = d.MoveTo(col, row)
		} else {
			row := d.Row + 1
			col := t.largestInRow(row, label) + 1
			d = d.MoveTo(col, row)
		}
	}

	return t.placeDomino(d)
}
```

Wait — re-reading the Java shuffle more carefully. The twist logic in the Java
code is:

```java
if (current.getIsVertical()) {
    current.flipDomino();  // now horizontal
    current.moveDomino(largestInRow(row + 1, label) + 1, row + 1);
} else {
    current.flipDomino();  // now vertical
    current.moveDomino(col + 1, largestInCol(col + 1, label) + 1);
}
```

So for a twist:
- If vertical: flip to horizontal, move to `(largestInRow(row+1, label)+1, row+1)`
- If horizontal: flip to vertical, move to `(col+1, largestInCol(col+1, label)+1)`

And for a slide:
- If vertical: move to `(col+1, largestInCol(col+1, label)+1)` — same orientation
- If horizontal: move to `(largestInRow(row+1, label)+1, row+1)` — same orientation

Note that `row` and `col` refer to the FIRST block's coordinates BEFORE the
flip/move. The corrected Go shuffle:

```go
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
```

Grid helpers:

```go
// largestInRow returns the largest column index in the given row
// occupied by a domino with label < bound.
func (t Tableau) largestInRow(row, bound int) int {
	maxCol := 0
	for col := 1; col < len(t.grid); col++ {
		if row < len(t.grid[col]) && t.grid[col][row] > 0 && t.grid[col][row] < bound {
			if col > maxCol {
				maxCol = col
			}
		}
	}
	return maxCol
}

// largestInCol returns the largest row index in the given column
// occupied by a domino with label < bound.
func (t Tableau) largestInCol(col, bound int) int {
	maxRow := 0
	if col < len(t.grid) {
		for row := 1; row < len(t.grid[col]); row++ {
			if t.grid[col][row] > 0 && t.grid[col][row] < bound {
				if row > maxRow {
					maxRow = row
				}
			}
		}
	}
	return maxRow
}

// overlapCount counts how many cells of domino `label` are also
// occupied by other dominoes.
func (t Tableau) overlapCount(label int) int {
	d := t.dominoes[label-1]
	count := 0
	if t.cellOccupiedByOther(d.Col, d.Row, label) {
		count++
	}
	if t.cellOccupiedByOther(d.SecondCol(), d.SecondRow(), label) {
		count++
	}
	return count
}
```

This is the main performance improvement over the Java version: `overlapCount`
is O(1) instead of O(n).

Right tableau and left tableau:

```go
func RightTableau(elem coxeter.Element) Tableau {
	return New(elem)
}

func LeftTableau(elem coxeter.Element) Tableau {
	return New(elem.Inverse())
}
```

Accessor methods:

| Method | Signature | Description |
|--------|-----------|-------------|
| Rank | `Rank() int` | Maximum number of dominoes |
| Size | `Size() int` | Number of dominoes currently placed |
| GetDomino | `GetDomino(label int) (Domino, bool)` | Get domino by label |
| MaxWidth | `MaxWidth() int` | Rightmost column |
| MaxHeight | `MaxHeight() int` | Bottommost row |
| Dominoes | `Dominoes() []Domino` | All placed dominoes |
| Equal | `Equal(other Tableau) bool` | Structural equality |

### heap.go

A heap visualizes a reduced expression by dropping blocks onto the Coxeter
graph.

```go
type Heap struct {
	blocks []Domino // one block per generator in the expression
	rank   int
	width  int
	height int
}
```

Constructor from a coxeter.Element:

```go
func NewHeap(elem coxeter.Element) Heap {
	expr := elem.ReducedExpression()
	length := expr.Length()

	blocks := make([]Domino, length)
	heights := make([]int, expr.Rank()+1) // heights[col]
	firstBlockUsed := make([]bool, length+1)
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
			row = max(heights[col], heights[col+1])
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
```

### tableau_test.go

Port test cases from `TableauTest.java`:

```go
func TestTableau_Construction(t *testing.T) {
	// Two elements that should produce the same right tableau
	w4, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	v4, _ := coxeter.NewElement([]int{1, 3, -4, -2})
	assert.True(t, tableau.New(w4).Equal(tableau.New(v4)))

	// Identity produces different tableau
	u4, _ := coxeter.NewElement([]int{1, 2, 3, 4})
	assert.False(t, tableau.New(w4).Equal(tableau.New(u4)))

	// Different ranks
	w46, _ := coxeter.NewElement([]int{1, -4, 3, -2, 5, 6})
	assert.False(t, tableau.New(w4).Equal(tableau.New(w46)))

	// Another pair with same tableau
	w6, _ := coxeter.NewElement([]int{-1, -6, 3, -4, 5, -2})
	v6, _ := coxeter.NewElement([]int{-6, -1, 3, -4, -2, 5})
	assert.True(t, tableau.New(w6).Equal(tableau.New(v6)))
}

func TestTableau_Dimensions(t *testing.T) {
	w4, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	tw4 := tableau.New(w4)
	assert.Equal(t, 4, tw4.MaxWidth())
	assert.Equal(t, 3, tw4.MaxHeight())

	w6, _ := coxeter.NewElement([]int{-1, -6, 3, -4, 5, -2})
	tw6 := tableau.New(w6)
	assert.Equal(t, 5, tw6.MaxWidth())
	assert.Equal(t, 4, tw6.MaxHeight())
}

func TestTableau_GetDomino(t *testing.T) {
	w4, _ := coxeter.NewElement([]int{1, -4, 3, -2})
	tw4 := tableau.New(w4)

	d, ok := tw4.GetDomino(4)
	assert.True(t, ok)
	assert.Equal(t, 2, d.Col)
	assert.Equal(t, 2, d.Row)
	assert.True(t, d.IsVertical)
}
```

## Phase 3: TikZ output — `internal/tikz`

Rendering is completely separate from data structures. Each function takes a
Tableau or Heap and returns a string.

```go
package tikz

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

func RenderHeap(h tableau.Heap) string {
	// Similar: iterate blocks, handle label 1 and 2 specially
	// with phantom spacing for the branching node
}
```

## Phase 4: CLI — `cmd/domino`

Subcommand-style CLI using only the standard library (`os.Args` parsing or a
small flag-based approach).

### Commands

| Command | Usage | Description |
|---------|-------|-------------|
| `info` | `domino info -perm 1,-4,3,-2` | Print length, descents, bad status, reduced expression |
| `tableau` | `domino tableau -perm 1,-4,3,-2` | Print right and left tableaux as TikZ |
| `heap` | `domino heap -perm 1,-4,3,-2` | Print heap as TikZ |
| `bad` | `domino bad -rank 4` | List all bad elements of D_n |

### Input format

Signed permutation as comma-separated integers: `-perm 1,-4,3,-2`
Generator expression as comma-separated integers: `-expr 1,2,3,4,3 -rank 4`

### Example output

```
$ domino info -perm 1,-4,3,-2
Permutation: [1, -4, 3, -2]
Length:       7
Right descent: {1, 2, 4}
Left descent:  {1, 2, 4}
Bad:           true
Reduced:       (4, 2, 1, 4, 3, 2, 4)

$ domino bad -rank 4
[1, -4, 3, -2]
[-3, 2, -1, 4]
...
(N bad elements in D_4)
```

### Implementation

```go
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/owner/domino/internal/coxeter"
	"github.com/owner/domino/internal/tableau"
	"github.com/owner/domino/internal/tikz"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "info":
		runInfo(os.Args[2:])
	case "tableau":
		runTableau(os.Args[2:])
	case "heap":
		runHeap(os.Args[2:])
	case "bad":
		runBad(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}
```

## Phase 5: Testing strategy

Every package gets thorough test coverage. Tests are the primary way we verify
mathematical correctness. Port all test vectors from the Java suite, then add
more.

- `testify/assert` for assertions, standard library for everything else
- Table-driven tests throughout
- `go test ./...` runs everything

### internal/coxeter tests

**TestElement_NewElement** — validation:
- Valid identity, valid even-negatives, valid with all signs
- Invalid: odd negatives, zero, out of range, repeated absolute value, empty

**TestElement_MapsTo** — signed permutation queries:
- Positive input, negative input, boundary values
- Port all `testMapsTo` and `testMapsFrom` cases from Java

**TestElement_Inverse** — algebraic properties:
- `w.Inverse().Inverse() == w` for multiple elements
- `identity.Inverse() == identity`
- Self-inverse elements (involutions): `[1, -4, 3, -2]` is its own inverse
- Port `testFindInverse` cases

**TestElement_RightMultiply** — group operation:
- `w * identity == w`
- `identity * w == w`
- `w * w_inverse == identity`
- Associativity: `(a * b) * c == a * (b * c)` for several triples
- Known products from Java test suite

**TestElement_Length** — Coxeter length:
- Identity has length 0
- Each generator has length 1
- Known lengths from Java: `[1,-4,3,-2]` → 7, `[1,3,2,4]` → 1, `[1,3,-4,-2]` → 8
- Length is invariant: `w.Length() == w.Inverse().Length()`

**TestElement_IsRightDescent** — descent tests:
- Identity has no descents
- Known descent sets from Java: `[1,-4,3,-2]` → {1,2,4}, `[1,3,2,4]` → {3}
- `w.LeftDescentSet() == w.Inverse().RightDescentSet()`

**TestElement_ReducedExpression** — round-trip:
- `w.ReducedExpression().ToElement() == w` for every test element
- `w.ReducedExpression().Length() == w.Length()`
- The expression is actually reduced: converting back and re-reducing gives same length

**TestElement_IsBad** — bad element detection:
- Known bad elements from thesis: `[1,-4,3,-2]`, `[-1,-6,3,-4,5,-2]`, `[1,-8,3,-6,5,-4,7,-2]`
- Known non-bad elements: identity, `[1,3,-4,-2]`, `[1,3,2,4]`
- Commuting generator products are not bad: `[-1,3,4,-2]`

**TestBadElements** — enumeration:
- D_3 has no bad elements
- D_4: verify count, verify every returned element passes `IsBad()`, verify
  known bad element `[1,-4,3,-2]` is in the list
- D_5: verify count, cross-check a sample
- D_6: verify count matches D_4 and D_5 trend (if feasible in test time)
- Verify no non-bad elements sneak in: for D_4, iterate ALL elements and confirm
  `BadElements(4)` matches the full brute-force filter

**TestBadElements_Optimized** — correctness of pruning:
- For D_4 and D_5, compare output of optimized `BadElements` against brute-force
  enumeration to ensure pruning doesn't miss any bad elements

**TestExpression_ToElement** — expression conversion:
- Known expressions from Java: applying generators to identity gives expected permutation
- Empty expression gives identity

**TestExpression_IsReduced** — reduction check:
- A reduced expression is reduced
- A non-reduced expression (e.g., `[3, 3]` which is s3*s3 = identity) is not

### internal/tableau tests

**TestTableau_Construction** — Garfinkle algorithm:
- Same-cell pairs from Java: `[1,-4,3,-2]` and `[1,3,-4,-2]` produce equal tableaux
- Different tableaux for different cells
- Same-cell pair at rank 6: `[-1,-6,3,-4,5,-2]` and `[-6,-1,3,-4,-2,5]`

**TestTableau_Dimensions** — width and height:
- Port all `testMaxWidth` and `testMaxHeight` from Java
- `[1,-4,3,-2]`: width=4, height=3
- `[-1,-6,3,-4,5,-2]`: width=5, height=4
- `[1,-4,3,-2,5,6]`: width=8, height=3

**TestTableau_GetDomino** — individual domino positions:
- `[1,-4,3,-2]` domino 4: col=2, row=2, vertical
- `[-1,-6,3,-4,5,-2]` domino 5: col=4, row=1, horizontal

**TestTableau_Identity** — identity element:
- All dominoes horizontal in first row
- Width = 2*rank, height = 1

**TestTableau_RightAndLeft** — left vs right tableau:
- Right tableau of w equals right tableau of any element in w's right cell
- Left tableau of w equals right tableau of w^{-1}

**TestTableau_Exhaustive_D4** — brute force for small rank:
- For every element of D_4 (192 elements), verify:
  - Tableau construction doesn't panic
  - Right tableau of w equals right tableau of w' iff they're in the same right cell
  - `RightTableau(w) == RightTableau(w')` implies `w` and `w'` have same right descent set

**TestHeap_Construction** — heap from element:
- Port `HeapTest.java` cases: verify height and width
- Heap size equals element length
- Heap of identity is empty

### internal/tikz tests

**TestRenderTableau** — TikZ output:
- Contains `\begin{tikzpicture}` and `\end{tikzpicture}`
- Contains correct number of `\node` lines (one per domino)
- Vertical dominoes use `[ver]`, horizontal use `[hor]`
- Specific coordinate values for known tableau

**TestRenderHeap** — TikZ output:
- Contains correct structure
- Labels 1 and 2 get phantom spacing
- Other labels render directly

### Verification approach

1. **Round-trip**: `ReducedExpression().ToElement()` equals the original for all test elements
2. **Same-cell invariant**: elements with same right tableau have same right descent set
3. **Cross-validation with Java**: for D_4, run all 192 elements through both Java and Go, compare tableau output
4. **Bad element correctness**: optimized enumeration matches brute-force for D_4 and D_5
5. **Algebraic identities**: `(w * v).Inverse() == v.Inverse() * w.Inverse()`, `w.Length() == w.Inverse().Length()`

## TODO

### Phase 0: Project setup
- [x] Initialize Go module (`go mod init`)
- [x] Create directory structure: `cmd/domino/`, `internal/coxeter/`, `internal/tableau/`, `internal/tikz/`
- [x] Add `testify` dependency (`go get github.com/stretchr/testify`)
- [x] Verify `go test ./...` runs (empty, no errors)

### Phase 1: `internal/coxeter/element.go`
- [x] Define `Element` struct with unexported `perm []int`
- [x] Implement `NewElement(perm []int) (Element, error)` — validate input (range, uniqueness, even negatives)
- [x] Implement `NewIdentity(rank int) Element`
- [x] Implement `Rank() int`
- [x] Implement `MapsTo(i int) int` — supports negative i
- [x] Implement `Equal(other Element) bool`
- [x] Implement `String() string` — one-line notation `[1, -4, 3, -2]`
- [x] Implement `Inverse() Element`
- [x] Implement `rightMultiplyGenerator(s int) Element` (unexported)
- [x] Implement `RightMultiply(other Element) Element`
- [x] Implement `LeftMultiply(other Element) Element`
- [x] Implement `countInversions(perm []int, factor int) int` (unexported)
- [x] Implement `Length() int`
- [x] Implement `IsRightDescent(s int) bool`
- [x] Implement `RightDescentSet() []int`
- [x] Implement `LeftDescentSet() []int`
- [x] Implement `isRightBad() bool` (unexported)
- [x] Implement `isCommutingProduct() bool` (unexported)
- [x] Implement `IsBad() bool`
- [x] Write `TestElement_NewElement` — valid and invalid inputs
- [x] Write `TestElement_MapsTo` — positive, negative, boundary
- [x] Write `TestElement_Inverse` — round-trip, involutions, identity
- [x] Write `TestElement_RightMultiply` — identity, inverse, associativity, known products
- [x] Write `TestElement_Length` — identity=0, generators=1, known values from Java
- [x] Write `TestElement_IsRightDescent` — identity has none, known descent sets
- [x] Write `TestElement_LeftDescentSet` — equals inverse's right descent set
- [x] Write `TestElement_IsBad` — known bad/non-bad elements from thesis
- [x] Run tests, verify all pass

### Phase 2: `internal/coxeter/expression.go`
- [x] Define `Expression` struct with unexported `generators []int` and `rank int`
- [x] Implement `NewExpression(generators []int, rank int) (Expression, error)` — validate generator range
- [x] Implement `Generators() []int` — returns copy
- [x] Implement `Rank() int`
- [x] Implement `Length() int`
- [x] Implement `String() string` — format as `(1, 3, 4, 3)`
- [x] Implement `ToElement() Element` — apply generators left to right from identity
- [x] Implement `IsReduced() bool`
- [x] Add `ReducedExpression() Expression` method to `Element`
- [x] Write `TestExpression_ToElement` — known expressions from Java
- [x] Write `TestExpression_IsReduced` — reduced and non-reduced expressions
- [x] Write `TestElement_ReducedExpression` — round-trip for all test elements, length matches
- [x] Run tests, verify all pass

### Phase 3: `internal/coxeter/bad.go` — bad element enumeration
- [x] Implement `allElements(rank int) []Element` — brute-force generation for testing
- [x] Implement `BadElements(rank int) []Element` — optimized backtracking
- [x] Implement `canBeRightBadPartial(perm []int, filledUpTo int) bool` — pruning on partial permutation
- [x] Implement inverse tracking — maintain partial inverse during backtracking
- [x] Implement even-parity constraint — enforce at last position
- [x] Implement parallel search — split on first position across goroutines
- [x] Write `TestBadElements_D3` — no bad elements
- [x] Write `TestBadElements_D4` — verify count, verify known element present, verify all returned are bad
- [x] Write `TestBadElements_D5` — verify count, spot check
- [x] Write `TestBadElements_BruteForceComparison` — for D_4, compare optimized vs brute-force on all 192 elements
- [x] Write `TestBadElements_BruteForceComparison_D5` — same for D_5 (1920 elements)
- [x] Run tests, verify all pass

### Phase 4: `internal/tableau/domino.go`
- [x] Define `Domino` struct — `Label`, `Col`, `Row`, `IsVertical` (exported fields)
- [x] Implement `SecondCol() int`
- [x] Implement `SecondRow() int`
- [x] Implement `MoveTo(col, row int) Domino`
- [x] Implement `Flip() Domino`
- [x] Implement `Equal(other Domino) bool`
- [x] Implement `String() string`
- [x] Write `TestDomino_SecondBlock` — vertical and horizontal cases
- [x] Write `TestDomino_MoveTo` — preserves label and orientation
- [x] Write `TestDomino_Flip` — swaps orientation, preserves position and label
- [x] Run tests, verify all pass

### Phase 5: `internal/tableau/tableau.go` — Garfinkle algorithm
- [x] Define `Tableau` struct — `rank`, `dominoes []Domino`, `present []bool`, `grid [][]int`
- [x] Implement `makeGrid(rank int) [][]int` — allocate 2D grid
- [x] Implement `placeDomino(d Domino) Tableau` — add domino to both slice and grid
- [x] Implement `removeDomino(label int) Tableau` — remove from both slice and grid
- [x] Implement `largestInRow(row, bound int) int` — max column in row with label < bound
- [x] Implement `largestInCol(col, bound int) int` — max row in column with label < bound
- [x] Implement `cellOccupiedByOther(col, row, label int) bool` — grid lookup
- [x] Implement `overlapCount(label int) int` — O(1) using grid
- [x] Implement `hasLarger(label int) bool` — any present domino with label > given
- [x] Implement `shuffle(label int) Tableau` — twist/slide logic from Java
- [x] Implement `addDomino(d Domino) Tableau` — Garfinkle alpha-map + shuffle
- [x] Implement `New(elem coxeter.Element) Tableau` — construct from element
- [x] Implement `RightTableau(elem coxeter.Element) Tableau`
- [x] Implement `LeftTableau(elem coxeter.Element) Tableau`
- [x] Implement `Rank() int`, `Size() int`, `MaxWidth() int`, `MaxHeight() int`
- [x] Implement `GetDomino(label int) (Domino, bool)`
- [x] Implement `Dominoes() []Domino` — all placed dominoes in label order
- [x] Implement `Equal(other Tableau) bool`
- [x] Write `TestTableau_Construction` — same-cell pairs produce equal tableaux
- [x] Write `TestTableau_Dimensions` — width/height for known elements
- [x] Write `TestTableau_GetDomino` — specific domino positions from Java tests
- [x] Write `TestTableau_Identity` — all horizontal in first row
- [x] Write `TestTableau_RightAndLeft` — left tableau equals right tableau of inverse
- [x] Write `TestTableau_Exhaustive_D4` — all 192 elements: no panics, descent set invariant
- [x] Run tests, verify all pass

### Phase 6: `internal/tableau/heap.go`
- [x] Define `Heap` struct — `blocks []Domino`, `rank`, `width`, `height`
- [x] Implement `NewHeap(elem coxeter.Element) Heap` — drop blocks algorithm
- [x] Implement `Blocks() []Domino`, `Rank() int`, `Size() int`, `MaxWidth() int`, `MaxHeight() int`
- [x] Write `TestHeap_Construction` — port HeapTest.java cases (height, width)
- [x] Write `TestHeap_Size` — heap size equals element length
- [x] Write `TestHeap_Identity` — identity produces empty heap
- [x] Run tests, verify all pass

### Phase 7: `internal/tikz/`
- [x] Implement `RenderTableau(t tableau.Tableau) string`
- [x] Implement `RenderHeap(h tableau.Heap) string` — handle labels 1/2 phantom spacing
- [x] Write `TestRenderTableau` — structure, node count, ver/hor styles, coordinates
- [x] Write `TestRenderHeap` — structure, phantom spacing for labels 1 and 2
- [x] Run tests, verify all pass

### Phase 8: `cmd/domino/main.go` — CLI
- [x] Implement argument parsing — subcommand dispatch
- [x] Implement `parsePerm(s string) ([]int, error)` — parse comma-separated ints
- [x] Implement `parseExpr(s string, rank int) (coxeter.Expression, error)`
- [x] Implement `runInfo(args []string)` — print length, descents, bad, reduced expression
- [x] Implement `runTableau(args []string)` — print right and left tableaux as TikZ
- [x] Implement `runHeap(args []string)` — print heap as TikZ
- [x] Implement `runBad(args []string)` — enumerate and print all bad elements for given rank
- [x] Implement `usage()` — help text
- [x] Build and manually test each subcommand
- [x] Verify `go build ./cmd/domino` produces working binary

### Phase 9: Cross-validation
- [x] For all 192 elements of D_4, compare Go tableau output against Java output
- [x] For all 192 elements of D_4, compare Go heap dimensions against Java output
- [x] For known bad elements from the thesis, verify Go and Java agree
- [x] Run `domino bad -rank 4` and `domino bad -rank 5`, verify output is plausible
- [x] Run full test suite: `go test ./...` — all green
