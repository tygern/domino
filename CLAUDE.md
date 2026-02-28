# CLAUDE.md

## Go Style

- Standard library preferred over frameworks
- Use Go 1.26+
- `go test ./...` to run tests

## Project Structure

```
cmd/           # Entry points — one subdirectory per binary
internal/      # Private application packages
pkg/           # Reusable public packages
```

## Naming

- **Packages**: short, lowercase, single word
- **Constructors**: always prefixed with `New` (`New`, `NewService`, `NewClient`)
- **Records/models**: plain structs, named for what they represent (`ChunkRecord`, `DataRecord`)

## Interfaces

- Small and focused — one or two methods
- Defined where they are consumed, not where they are implemented
- Unexported when only used within one package

## Dependency Injection

- Constructor-based, no DI frameworks
- Interfaces for external dependencies, concrete types for internal ones
- No global state — everything wired through constructors in `main()`

## Error Handling

- Wrap with context: `fmt.Errorf("unable to list ids: %w", err)`
- Combine multiple: `errors.Join(errs...)`
- Check with `errors.Is()`
- No custom error types — use standard `error` and wrapping
- Explicitly ignore with `_ = thing.Close()`

## Logging

- `log/slog` — structured key-value pairs
- Minimal — log at boundaries, not every step

## Configuration

- Environment variables, fail fast on missing required values at startup

## Database

- `database/sql` — no ORMs
- Write generic helpers for scanning rows into records
- Raw SQL for migrations

## Concurrency

- `context.Context` passed through for cancellation and timeouts
- Goroutines + channels when needed
- Buffered channels when producer shouldn't block

## Testing

- Build your own assertions with the standard library
- Test function naming: `TestTypeName_MethodName`
- Hand-written fakes over mock libraries
- `t.Cleanup()` for teardown
- Isolate external dependencies per test (test DBs, ephemeral servers)

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
7. Interfaces (if defined locally)

## General Principles

- Minimal comments — comment the "why" not the "what"
- No unnecessary abstractions — three similar lines is better than a premature helper
- Each package has a clear single responsibility
- Prefer closures for injecting dependencies into function values
