# AGENTS.md — Go Code Quality Guidelines

This file governs how AI agents (and human contributors) write, review, and modify Go code in this repository. All rules apply unless a specific file or package has a documented exception.

---

## Table of Contents

1. [Core Philosophy](#core-philosophy)
2. [Tooling & Formatting](#tooling--formatting)
3. [Naming Conventions](#naming-conventions)
4. [Package Design](#package-design)
5. [Error Handling](#error-handling)
6. [Concurrency](#concurrency)
7. [Interfaces & Abstractions](#interfaces--abstractions)
8. [Testing](#testing)
9. [Performance](#performance)
10. [Security](#security)
11. [Documentation](#documentation)
12. [Dependency Management](#dependency-management)
13. [Code Review Checklist](#code-review-checklist)

---

## Core Philosophy

- Prefer **clarity over cleverness**. Code is read far more than it is written.
- Prefer **explicit over implicit**. Avoid magic; make data flow and control flow obvious.
- Follow the [Go Proverbs](https://go-proverbs.github.io/). When in doubt, ask: _"What would the standard library do?"_
- **DO NOT** introduce abstractions preemptively. Solve the concrete problem first.
- **DO NOT** port idioms from other languages (OOP hierarchies, functional pipelines, etc.) unless they are the clearest Go solution.

---

## Tooling & Formatting

### Mandatory Tools

- **MUST** run `gofmt` (or `goimports`) before every commit. Zero tolerance for unformatted code.
- **MUST** run `go vet ./...` and fix all reported issues before committing.
<!-- - **MUST** run `staticcheck ./...` and resolve all findings unless explicitly annotated with `//nolint:...` and a justification comment. -->
<!-- - **MUST** use `golangci-lint` with the repository's `.golangci.yml` configuration. Do not add new linter exclusions without team review. -->
<!-- - should run `govulncheck ./...` before releases to detect known vulnerabilities in dependencies. -->

### Formatting Rules

- **MUST** use tabs for indentation, never spaces — this is enforced by `gofmt`.
- **MUST** keep lines under **120 characters** where practical; `gofmt` will not enforce this, but reviewers will flag egregious violations.
- **MUST NOT** add blank lines at the start or end of function bodies.
- should group imports as: (1) standard library, (2) third-party, (3) internal packages — separated by blank lines. `goimports` does this automatically.
- should not mix declaration styles (`var`, `:=`) arbitrarily within a single function; prefer `:=` for local short-lived variables and `var` for zero-value declarations or multi-line declarations.

---

## Naming Conventions

- **MUST** follow the [Effective Go naming guidelines](https://go.dev/doc/effective_go#names).
- **MUST** use `MixedCaps` or `mixedCaps` — never underscores in identifier names (except in generated code and test helper files).
- **MUST NOT** use single-letter variable names outside of loop indices (`i`, `j`, `k`) and very short mathematical functions where the single letter has an obvious domain meaning.
- **MUST NOT** stutter package names into exported identifiers (e.g., `http.HTTPServer` → wrong; `http.Server` → correct).
- **MUST NOT** name variables `err2`, `err3`, etc. Use descriptive names: `parseErr`, `closeErr`.
- should keep acronyms consistently cased: `userID`, `parseURL`, `serveHTTP`, `writeJSON` — all letters of the acronym uppercase.
- should name boolean variables and functions so they read naturally as a question: `isReady`, `hasChildren`, `canRetry`, `IsValid()`.
- should name interfaces with a single method after that method plus `-er`: `Reader`, `Writer`, `Stringer`. Multi-method interfaces should have a descriptive noun name.
- should avoid naming a return variable `result` or `res` — prefer domain-specific names.
- Context variables should always be named `ctx`. Never `c`, `context`, or anything else.

---

## Package Design

- **MUST** design packages around _what they provide_, not _what they contain_. A package called `util`, `common`, `helpers`, or `misc` is a design smell and **MUST NOT** be introduced.
- **MUST NOT** create circular dependencies between packages.
- **MUST** keep `internal/` packages for code that should not be imported by external consumers.
- **MUST NOT** export a type or function unless it is part of the public API. When in doubt, keep it unexported.
- should aim for packages with a single, clearly articulable responsibility.
- should avoid `init()` functions. If initialization is required, provide an explicit constructor or `Open`/`New` function that the caller invokes.
- should prefer flat package hierarchies. Deep nesting (e.g., `pkg/a/b/c/d`) is usually a sign of over-engineering.
- should not place `main` logic in `internal/` — `main` packages belong in `cmd/<name>/main.go`.

---

## Error Handling

- **MUST** handle every error. **Never** assign an error to `_` unless there is an explicit, commented reason.
- **MUST** wrap errors with context using `fmt.Errorf("doing X: %w", err)` so that callers can inspect the chain.
- **MUST NOT** use `panic` for ordinary error conditions. `panic` is reserved for unrecoverable programmer errors (e.g., nil pointer dereference in an invariant that should never be violated). Library code **MUST NOT** panic at all.
- **MUST NOT** log an error and also return it — choose one. Logging and returning causes double-logging throughout call stacks.
- **MUST** use sentinel errors (`var ErrNotFound = errors.New(...)`) or custom error types for errors that callers are expected to inspect. Use `errors.Is` and `errors.As` for comparisons — never `==` on wrapped errors.
- should define custom error types when errors carry structured data (e.g., an HTTP status code, a field name, an offset).
- should place error returns last in the return value list.
- should not use boolean return values to signal success/failure when an `error` would convey more information.
- should not use error strings as the sole means of distinguishing error kinds.

```go
// BAD
if err != nil && err.Error() == "not found" { ... }

// GOOD
if errors.Is(err, ErrNotFound) { ... }
```

---

## Concurrency

- **MUST** document every goroutine's lifetime: when it starts, what causes it to stop, and who is responsible for cleanup.
- **MUST NOT** leak goroutines. Every goroutine **MUST** have a defined exit path, typically via a `context.Context` cancellation or a `done` channel.
- **MUST** pass `context.Context` as the first argument to any function that performs I/O, blocking operations, or spawns goroutines.
- **MUST NOT** store a `context.Context` in a struct field unless there is a compelling architectural reason, documented in a comment.
- **MUST** protect all shared mutable state with a `sync.Mutex`, `sync.RWMutex`, or equivalent. Annotate the struct field it protects with a comment: `// mu protects the fields below`.
- **MUST** run tests with `-race` in CI. The race detector is non-negotiable.
- **MUST NOT** use `time.Sleep` to synchronize goroutines in production code. Use channels or synchronization primitives.
- should prefer channels for ownership transfer and `sync` primitives for guarding shared state.
- should avoid `sync.Map` unless profiling shows it outperforms a mutex-protected map in the specific workload. It is harder to use correctly.
- should use `errgroup.Group` (from `golang.org/x/sync/errgroup`) when fanning out goroutines and collecting errors.
- should close channels from the sender side only, never the receiver side.

---

## Interfaces & Abstractions

- **MUST** accept interfaces, return concrete types — this is the Go way. Functions should take the narrowest interface sufficient for their needs.
- **MUST NOT** define an interface for a type that has only one concrete implementation unless: (a) testing requires a mock, or (b) a second implementation is planned within the current milestone.
- **MUST NOT** use empty interfaces (`interface{}` or `any`) unless absolutely necessary. Prefer generics (Go 1.18+) or concrete types.
- **MUST** keep interfaces small. An interface with more than five methods is usually doing too much and should be split.
- should define interfaces in the _consuming_ package, not the _providing_ package.
- should not embed interfaces in structs as a way to satisfy them partially ("embedding hack"). Implement all methods explicitly.

---

## Testing

- **MUST** write tests for all exported functions and methods. Minimum line coverage of **80%** is enforced in CI; new code below this threshold will block merging.
- **MUST** use table-driven tests for functions with multiple input/output cases.
- **MUST** name test cases descriptively in the `name` field of each table entry; a failing test's name must tell you what broke without reading the test body.
- **MUST NOT** use `time.Sleep` in tests. Use test helpers, fake clocks, or channels to synchronize.
- **MUST NOT** test implementation details — test behavior and public contracts.
- **MUST** use `t.Helper()` in test helper functions so failure lines point to the caller, not the helper.
- **MUST** clean up resources in tests using `t.Cleanup(func() { ... })`, not deferred functions in setup code.
- should use `testify/assert` and `testify/require` for readability, with `require` for fatal assertions and `assert` for non-fatal ones.
- should use subtests (`t.Run(...)`) for logical grouping; this enables running individual cases with `-run`.
- should keep test files in the same package as the code under test (`package foo`) for whitebox testing. Use `package foo_test` for blackbox integration tests.
- should use `testing/iotest`, `net/http/httptest`, and other standard test helpers before reaching for third-party fakes.
- should benchmark performance-sensitive code with `testing.B` and commit baseline benchmarks to `testdata/` or a `bench_test.go` file.

```go
// Table-driven test structure (canonical form)
func TestMyFunc(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {name: "empty string returns empty", input: "", want: ""},
        {name: "trims leading whitespace", input: "  foo", want: "foo"},
        {name: "invalid utf8 returns error", input: "\xff", wantErr: true},
    }
    for _, tc := range tests {
        tc := tc // capture loop variable (pre-Go 1.22)
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            got, err := MyFunc(tc.input)
            if tc.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tc.want, got)
        })
    }
}
```

---

## Performance

- **MUST NOT** optimize prematurely. Write clear, idiomatic code first. Profile before optimizing.
- **MUST** use `pprof` and benchmark tests (`testing.B`) to justify any non-obvious optimization.
- should pre-allocate slices and maps when the capacity is known: `make([]T, 0, n)`, `make(map[K]V, n)`.
- should prefer `strings.Builder` over concatenation in loops.
- should use `sync.Pool` for high-frequency short-lived allocations that are profiler-confirmed hotspots.
- should avoid unnecessary conversions between `[]byte` and `string` in hot paths.
- should prefer passing structs by pointer when they are larger than ~64 bytes or contain synchronization primitives.
- should profile allocations with `go test -bench=. -benchmem` before claiming a change is a performance improvement.

---

## Security

- **MUST NOT** log secrets, passwords, tokens, or PII. Use structured logging and explicitly filter sensitive fields.
- **MUST NOT** construct SQL queries with string concatenation. **MUST** use parameterized queries or a query builder that safely handles parameters.
- **MUST NOT** use `math/rand` for security-sensitive randomness. **MUST** use `crypto/rand`.
- **MUST** validate and sanitize all external input (HTTP bodies, CLI args, environment variables, config files) at the boundary where it enters the system.
- **MUST** use `html/template` (not `text/template`) for HTML output to prevent XSS.
- **MUST NOT** suppress TLS certificate verification (`InsecureSkipVerify: true`) in production code. Test code that requires this **MUST** be gated by a build tag.
- **MUST** use `net/http`'s `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` on all HTTP servers.
- should use `golang.org/x/crypto` for modern cryptographic algorithms not yet in the standard library.
- should run `govulncheck` as part of the release pipeline.

---

## Documentation

- **MUST** write a godoc comment for every exported identifier. The comment **MUST** begin with the name of the identifier.
- **MUST** document the concurrency safety of exported types that contain synchronization: _"Foo is safe for concurrent use."_ or _"Foo must not be used concurrently."_
- **MUST** document error conditions in function comments: which errors are returned, under what conditions, and whether they are sentinel errors that callers should inspect.
- **MUST** include a `// Deprecated: use X instead.` comment on deprecated functions before removal.
- should add package-level doc comments in `doc.go` for packages with non-trivial APIs.
- should include runnable `Example` functions in `_test.go` files for key public APIs; these appear in godoc and are verified by `go test`.
- should avoid commenting _what_ the code does (the code shows that); comment _why_ non-obvious decisions were made.

```go
// ParseDuration parses a duration string with an optional unit suffix.
// Supported units: "ms", "s", "m", "h". Unitless integers are treated as seconds.
//
// It returns ErrInvalidUnit if the suffix is not recognized, or a standard
// strconv error if the numeric part cannot be parsed.
func ParseDuration(s string) (time.Duration, error) { ... }
```

---

## Dependency Management

- **MUST** use Go modules (`go.mod` / `go.sum`) and commit both files.
- **MUST NOT** vendor dependencies unless there is an explicit policy decision to do so (e.g., air-gapped builds).
- **MUST NOT** import a new third-party dependency without adding a comment in the PR description explaining what it does and why the standard library is insufficient.
- **MUST** pin dependencies to specific versions; do not use pseudo-versions in `go.mod` for production dependencies unless pulling from a commit that predates tagging.
- should regularly run `go mod tidy` and commit the result; stale entries in `go.mod`/`go.sum` are a code smell.
- should prefer standard library solutions. Before adding a dependency, ask: _"Can I implement this in under 50 lines without sacrificing correctness?"_ If yes, do so.
- should audit licenses of new dependencies. AGPL and other copyleft licenses **MUST** be flagged for legal review before merging.

---

## Code Review Checklist

Before marking a PR ready for review, verify:

- [ ] `gofmt`/`goimports` has been run; output is unchanged
- [ ] `go vet ./...` passes with zero warnings
- [ ] `staticcheck ./...` passes or all suppressions are annotated and justified
- [ ] `go test -race ./...` passes
- [ ] New exported symbols have godoc comments
- [ ] Every new error is handled; no `_` suppression without a comment
- [ ] No new `util`, `common`, or `misc` packages introduced
- [ ] No secrets or sensitive data in logs or error messages
- [ ] No goroutines are spawned without a documented exit condition
- [ ] Dependencies added to `go.mod` are justified in the PR description
- [ ] Table-driven tests cover happy path, error path, and edge cases
- [ ] Benchmarks updated if a performance-sensitive path was changed

---

_Last updated: 2026-02-28. Raise a PR against this file to propose changes; all modifications require at least two approvals._
