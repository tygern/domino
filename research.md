# Research: BuildDomino Rewrite

## What the program does

BuildDomino constructs domino tableaux and heaps from elements of type D Coxeter groups. It implements the Garfinkle algorithm (a generalization of the Robinson-Schensted correspondence for signed permutations) and provides elementary group computations like length, descent sets, inverses, and bad element detection. Output is either textual, graphical (Swing), or TikZ code for LaTeX.

The thesis motivating this work is "Leading Coefficients of Kazhdan-Lusztig Polynomials in Type D" (arXiv:1304.6074). The core result: the leading coefficient mu(x,w) of a Kazhdan-Lusztig polynomial is 0 or 1 when x is fully commutative in type D, extending a known result from type A. The domino tableau machinery is the computational backbone for working with Kazhdan-Lusztig cells in type D.

## Mathematical background

### Coxeter groups

A Coxeter group W is defined by generators S = {s_1, ..., s_n} and relations (s_i s_j)^{m_{ij}} = e. The length l(w) of an element w is the minimum number of generators in any expression for w. Such a minimal expression is called a reduced expression. A right descent of w is a generator s such that l(ws) < l(w).

### Type D

The type D_n Coxeter group has generators s_1, s_2, ..., s_n with relations:
- s_i^2 = e for all i
- s_1 s_3 = s_3 s_1 (s_1 and s_3 commute — the branching)
- s_2 s_3 = s_3 s_2 (s_2 and s_3 commute — the other branch)
- s_i s_{i+1} s_i = s_{i+1} s_i s_{i+1} for adjacent non-branching generators
- s_i s_j = s_j s_i when |i-j| >= 2 (and not in the branch)

The Coxeter-Dynkin diagram for D_n has a fork: nodes s_1 and s_2 both connect to s_3, which connects linearly to s_4, ..., s_n. The existing code uses s_2 as the "branch node" — note this is a labeling convention that differs from some textbook sources.

Type D_n is realized as the subgroup of signed permutations (bijections w: {-n,...,-1,1,...,n} -> {-n,...,-1,1,...,n} with w(-i) = -w(i)) where the number of negative values in {w(1),...,w(n)} is even. This is an index-2 subgroup of the type B_n hyperoctahedral group (which allows any number of negatives).

In this realization, the generators act as:
- s_1: swaps positions 1 and 2 AND negates both (i.e., sends (a,b,...) to (-b,-a,...))
- s_i (i >= 2): swaps positions i-1 and i (standard adjacent transposition)

### Fully commutative and bad elements

An element w is fully commutative if any reduced expression for w can be obtained from any other using only commutation relations (s_i s_j = s_j s_i). A bad element is one that has no reduced expression beginning or ending in two noncommuting generators, and is not itself a product of commuting generators. Bad elements are the problematic cases for computing mu-values in type D.

### Kazhdan-Lusztig polynomials and cells

Kazhdan-Lusztig polynomials P_{x,w} are indexed by pairs of group elements and arise from the Hecke algebra of the Coxeter group. The leading coefficient mu(x,w) determines cell structure. Two elements are in the same left cell if they have the same left tableau under the Garfinkle correspondence; same right cell if they have the same right tableau. Computing cells requires moving tableaux through "cycles" (open and closed) — this is mentioned as future work in the README.

### Domino tableaux

A domino tableau is a Young-diagram-like shape filled with dominoes (1x2 or 2x1 tiles) labeled 1 through n. Each domino occupies two adjacent cells. The shape must be a valid Young diagram, and labels increase along rows and columns. Vertical dominoes have negative labels in the code; horizontal have positive.

The Garfinkle algorithm constructs a pair of tableaux (P, Q) from a signed permutation w:
- The right tableau P is built from w
- The left tableau Q is built from w^{-1}
- Construction proceeds by adding dominoes one at a time (label 1 through n)
- Each addition may trigger "shuffling" — a sequence of twists and slides that rearrange existing dominoes to maintain the tableau property

The shuffling step is the most complex part of the algorithm. For each new domino being inserted, larger-labeled dominoes may need to be moved. A twist flips a domino's orientation and shifts it; a slide moves it parallel to its orientation. The choice between twist and slide depends on overlap counts with neighboring dominoes.

### Heaps

A heap is a poset visualization of a reduced expression. Generators are represented as blocks dropped onto the Coxeter graph. Non-commuting generators stack; commuting generators sit side by side. For type D, the branch means s_1 and s_2 share a column conceptually but don't actually overlap, which is tricky to render in 2D.

## Existing codebase analysis

### Language and structure

Java, ~3400 lines across 16 source files and 9 test files. Built with Apache Buildr (Ruby-based). No external dependencies beyond JUnit and Java SE. Uses Swing for graphical output.

### Class hierarchy

```
Element (abstract) — signed permutation, inversions, descents
├── EvenElement (abstract) — adds countNeg()
│   └── TypeD — type D operations, generator multiplication, reduced expressions
└── TypeA — symmetric group (reference implementation)

Expression (abstract) — generator sequences
├── TypeDExpression
└── TypeAExpression

Tableau — array of Domino objects, Garfinkle shuffling algorithm
Heap — block diagram of reduced expression

Domino — label + two Coordinate pairs + orientation
Coordinate — (x, y) pair
BoundedSet — bounded integer set (for descent sets)

DrawDomino, DrawHeap — Swing rendering
BuildDomino — interactive REPL
AccessFile — resource loader
```

### Key algorithms and their complexity

| Operation | Method | Complexity | Notes |
|---|---|---|---|
| Length | `TypeD.length()` | O(n log n) | Merge-sort inversion counting |
| Reduced expression | `TypeD.findRE()` | O(L * n) | L = length, greedy right-to-left |
| Right descent | `TypeD.isRightDescent(s)` | O(1) | Position comparison |
| Descent set | `TypeD.rightDescent()` | O(n) | All generators checked |
| Bad element test | `TypeD.isBad()` | O(n) | Checks left-bad AND right-bad |
| Inverse | `Element.invertPermutation()` | O(n) | Array reversal |
| Multiply | `Element.rightMultiplyPerm()` | O(n) | Composition |
| Tableau construction | `Tableau(TypeD)` | O(n^3) worst | n dominoes, each may shuffle O(n) others |
| Heap construction | `Heap(TypeD)` | O(L) | Single pass through expression |

### Problems with the existing code

1. **Mutable everywhere.** Domino, Coordinate, Tableau all have setters and mutation methods. Shuffling mutates dominoes in place. This makes reasoning about correctness harder and prevents safe concurrency.

2. **No separation of concerns.** Tableau mixes construction algorithm with rendering (tikzDraw, screenDraw). Element mixes group algebra with I/O formatting.

3. **Inefficient data structures.** Tableau stores dominoes in a flat array and does linear scans for lookups by position. `largestInRow` and `largestInCol` are O(n) scans called repeatedly during shuffling.

4. **Type A is dead weight.** Only used for validation during development. Not needed in production.

5. **The REPL is tightly coupled.** BuildDomino.java is 326 lines of static methods parsing stdin. No clean API boundary.

6. **No batch mode.** Can only process one element at a time interactively.

7. **GUI code (Swing) is mixed in.** DrawDomino and DrawHeap are Java-specific rendering tied to the data model.

8. **Tests are thin.** 9 test files but low coverage of edge cases, especially around the shuffling algorithm.

9. **Reduced expression generation is quadratic.** `findRE()` iterates length times, each time scanning generators — could be improved.

10. **No support for cycles.** The README mentions Garfinkle's cycle operations (needed for computing Kazhdan-Lusztig cells) as future work. This is the main missing feature.

## Performance considerations for the Go rewrite

### What matters

The expensive operation is tableau construction, specifically the shuffling step. For large rank n, constructing a single tableau is O(n^3). If the goal is to compute cells (which involves many elements), we need this to be fast.

### Opportunities

1. **Immutable value types.** Go structs are value types by default. Represent Coordinate, Domino as small value types. Copy-on-write semantics are natural.

2. **Spatial lookup for dominoes.** Instead of a flat array with linear scans, use a 2D grid (slice of slices) indexed by position. O(1) lookup for "what domino is at (r,c)?" instead of O(n).

3. **Preallocated memory.** For rank n, we know the maximum tableau size. Preallocate slices instead of dynamically growing ArrayLists.

4. **Batch processing.** Design a library API, not a REPL. Enable processing many elements programmatically.

5. **Concurrency.** Left and right tableaux can be computed independently (different inputs: w and w^{-1}). Tableau computations for different elements are embarrassingly parallel.

6. **Avoid allocations in hot paths.** The shuffling inner loop should not allocate. Use index-based operations on preallocated grids.

7. **Inversion counting.** The merge-sort approach is already O(n log n). Keep it.

### What not to optimize

- Descent checks and generator multiplication are O(1) or O(n) and already fast.
- Heap construction is O(L) and not a bottleneck.
- Reduced expression generation: O(L * n) is fine for typical ranks.

## Proposed package structure

```
cmd/
  domino/           # CLI entry point

internal/
  coxeter/          # Type D group elements, expressions, generators, length, descents
  tableau/          # Domino tableau construction (Garfinkle algorithm)
  heap/             # Heap construction from reduced expressions
  tikz/             # TikZ output generation

pkg/
  domino/           # Domino, Coordinate value types (shared across packages)
```

### Key design decisions

1. **No GUI.** Drop Swing entirely. Keep TikZ output. Add JSON or plain text output for programmatic use.

2. **Library first.** The core packages should be importable and usable without the CLI. The CLI is a thin wrapper.

3. **Immutable core types.** Domino and Coordinate should be immutable structs. Tableau construction returns a new Tableau rather than mutating.

4. **Grid-based tableau.** Internally, the tableau should maintain a 2D grid for O(1) positional lookups alongside the domino list.

5. **Type D only.** Drop Type A entirely — it was only a reference implementation.

6. **Separate algorithm from rendering.** Tableau knows nothing about TikZ or display. Rendering is a separate package that reads tableau data.

## Open questions

1. **Cycles**: The README mentions cycle computations as future work. Should the Go rewrite include this? It's the main missing feature for computing Kazhdan-Lusztig cells, which is the whole point of the thesis.

2. **Type B**: The hyperoctahedral group (type B) is the parent group of type D. Should we support both? The signed permutation infrastructure is nearly identical — type D just adds the even-negatives constraint.

3. **Output formats**: TikZ is useful for papers. What about SVG for web? JSON for data exchange? Or is TikZ sufficient?

4. **Rank bounds**: The Java code has a hardcoded `rankBound = 3` default. What ranks do we need to support? Performance characteristics change significantly above rank ~20.

5. **Batch interface**: Should the CLI accept files of elements to process? Or is a Go library API sufficient for batch work?

## References

- Gern, "Leading Coefficients of Kazhdan-Lusztig Polynomials in Type D" — https://arxiv.org/abs/1304.6074
- Garfinkle, "On the classification of primitive ideals for complex classical Lie algebras" — Parts [I](http://www.numdam.org/item?id=CM_1990__75_2_135_0), [II](http://www.numdam.org/item?id=CM_1992__81_3_307_0), [III](http://www.numdam.org/item?id=CM_1993__88_2_187_0)
- Humphreys, "Reflection Groups and Coxeter Groups" — Cambridge University Press
- Stembridge, "On the Fully Commutative Elements of Coxeter Groups" — https://link.springer.com/article/10.1023/A:1022452717148
- BuildDomino source — https://github.com/tygern/BuildDomino
