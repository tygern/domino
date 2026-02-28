# CLAUDE.md

## Build & Test

- `go test ./...` to run all tests
- `go build ./cmd/domino` to build the CLI
- `go vet ./...` to check for issues
- Go 1.25+, module path `github.com/tygern/domino`

## Project Structure

```
cmd/domino/        # CLI entry point
internal/
  coxeter/         # Type D signed permutations, expressions, bad element enumeration
  tableau/         # Domino tableaux (Garfinkle algorithm), heaps
  tikz/            # TikZ rendering for tableaux and heaps
```

## Key Types

- `coxeter.Element` — signed permutation (immutable, value type). Constructor `NewElement([]int)` validates even-negatives constraint.
- `coxeter.Expression` — sequence of generators. Constructor `NewExpression([]int, rank)`. Convert with `ToElement()` / `elem.ReducedExpression()`.
- `tableau.Domino` — immutable value type with `Label`, `Col`, `Row`, `IsVertical`.
- `tableau.Tableau` — grid-backed domino tableau. Constructed via `tableau.New(elem)`, accessed via `RightTableau(elem)` / `LeftTableau(elem)`.
- `tableau.Heap` — heap poset. Constructed via `NewHeap(elem)`.

## Design Conventions

- All core types are immutable — methods return new values, never mutate
- `Tableau` uses an internal 2D grid for O(1) positional lookups during shuffling
- `BadElements(rank)` uses parallel backtracking with pruning, not brute force
- Rendering (tikz package) is fully separated from data (tableau package)

## Generator Numbering

Type D generators use the convention from the thesis (arXiv:1304.6074):
- `s_1`: swaps positions 1,2 with negation
- `s_2` through `s_n`: adjacent transpositions (s_i swaps positions i-1, i)
- `s_1` and `s_2` both connect to `s_3` (the branch)

## Go Style

- Standard library preferred over frameworks
- Only external dependency is `testify` for test assertions
- Minimal comments — comment the "why" not the "what"
- No unnecessary abstractions

## Naming

- **Packages**: short, lowercase, single word
- **Constructors**: prefixed with `New` (`NewElement`, `NewExpression`, `NewHeap`)
- **Test functions**: `TestTypeName_MethodName`

## Error Handling

- Constructors return `(T, error)`, validated at boundaries
- CLI exits with `fmt.Fprintln(os.Stderr, ...)` and `os.Exit(1)`
- Wrap with context: `fmt.Errorf("description: %w", err)`

## Imports

Two groups separated by a blank line:

1. Standard library (alphabetical)
2. Everything else (alphabetical)

## Code Organization Within Files

1. Package declaration
2. Imports
3. Type definitions
4. Constructor (`New*`)
5. Methods
6. Private helpers
