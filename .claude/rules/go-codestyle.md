---
applyTo: "**/*.go"
---

## Control Flow

- NEVER use combined assignment and condition in one statement. Always split into two lines:

  ```go
  // WRONG — never do this
  if err := doSomething(); err != nil { ... }

  // CORRECT — always split
  err := doSomething()
  if err != nil { ... }
  ```

  This applies to every case without exception — short calls, single-liners, all of them.

- don't use multi expressions in "if" and other statements. Split to lines

## Error Handling

- wrap return errors and provide context using `fmt.Errorf("...: %w", err)` to preserve the error chain
- don't use "failed" word in error messages. Describe the operation that went wrong:

  ```go
  // WRONG
  return fmt.Errorf("failed to read config: %w", err)

  // CORRECT
  return fmt.Errorf("read config: %w", err)
  ```

- use lowercase first letters in error messages (no capitalization)
- use `errors.Is()` and `errors.As()` for error comparison instead of direct comparison
- don't use named return variables (`nonamedreturns` linter is enabled)
- error is always the last return value

## Naming

- use user-friendly variable names; balance length and clarity for the developer
- don't use short names for function receiver names — use the full type name in lowercase:

  ```go
  // WRONG
  func (d *ComponentDeployer) Deploy() ...

  // CORRECT
  func (deployer *ComponentDeployer) Deploy() ...
  ```

- use singular names for packages (e.g., `handler` not `handlers`, `notify` not `notifications`)
- prefix interface names with behavior verb (e.g., `Reader`, `Writer`, `Validator`)
- use `New<Type>()` constructor pattern for structs that need initialization
- use `any` instead of `interface{}`

## HTTP

- use `http.Status*` constants for HTTP status codes (e.g., `http.StatusOK`, `http.StatusBadRequest`)
- use `http.Method*` constants for HTTP method names (e.g., `http.MethodPost`, `http.MethodGet`)

## Struct Visibility

- encapsulate struct members as much as possible — only what is used externally should be public
- config structs that need YAML/JSON unmarshaling: exported fields
- internal state structs: unexported fields with exported methods

## Function Signatures

- functions must accept `logrus.FieldLogger` logger as input parameter (not concrete `*logrus.Logger`)
- use `context.Context` where possible
- standard parameter order: `ctx context.Context`, `logger logrus.FieldLogger`, then domain objects
- keep functions short — if over 40-50 lines, consider splitting
- place exported functions before unexported ones in a file

## Logging

- alias logrus import as `log`: `log "github.com/sirupsen/logrus"`
- use lowercase first letters in logger messages:

  ```go
  log.Infof("starting server on port %s", port)
  ```

- use `log.WithError(err).Error("message")` to attach error context
- use appropriate log levels: `Debug` for tracing, `Info` for operational events, `Warn` for recoverable issues, `Error` for failures

## Imports

- group imports in three blocks separated by blank lines: stdlib, internal (`deployer/...`), external vendors
- order within each group: alphabetical

  ```go
  import (
      "context"
      "fmt"
      "net/http"

      "deployer/src/config"
      "deployer/src/core"

      log "github.com/sirupsen/logrus"
  )
  ```

## Concurrency

- always pass `context.Context` as the first parameter
- prefer channels for communication, mutexes for state protection
- use `defer` for cleanup (mutex unlock, file close, etc.)
- use `sync.WaitGroup` to wait for multiple goroutines; prefer `wg.Go(func(){...})` (Go 1.25+) over manual `wg.Add(1)` + `go func(){ defer wg.Done() }()` — the toolchain flags the manual form
- use `select` with `context.Done()` for cancellation-aware loops

## Testing

- use table-driven tests with descriptive test case names
- use `t.Helper()` in test helper functions
- use stdlib `testing` with `reflect.DeepEqual` for assertions — testify is not a dependency and must not be added (see Dependencies)
- use `defer` for test cleanup/teardown (e.g., restoring global state)
- iterate test cases with `for _, testCase := range tests` and `t.Run(testCase.name, ...)`

## Performance

- preallocate slices with `make([]T, 0, capacity)` when size is known
- use `strings.Builder` for string concatenation in loops

## Dependencies

- avoid adding new dependencies for things achievable with stdlib
- for terminal coloring use raw ANSI escape codes; do not add `fatih/color` or similar libraries

## Linter Compliance

The project uses golangci-lint v2 with these key linters — code must pass all of them:

- `nonamedreturns` — no named return variables (report-error-in-defer enabled)
- `errorlint` — use `%w` for wrapping, `errors.Is()`/`errors.As()` for comparison
- `gosec` — no security issues; use `//nolint:gosec` with justification comment when needed
- `godot` — end comments with a period
- `nestif` — avoid deeply nested if statements
- `cyclop` — keep cyclomatic complexity low
- `lll` — keep lines within length limit
- `prealloc` — preallocate slices when possible
- `usestdlibvars` — use stdlib constants (HTTP status codes, methods, etc.)
- `wsl_v5` — whitespace linter for consistent blank line usage
- `gochecknoinits` — no `init()` functions
- `bodyclose` — always close HTTP response bodies
- `inamedparam` — use named parameters in interfaces
