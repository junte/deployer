---
applyTo: "**/*.go"
---

- wrap return errors and provide context
- don't use named return variables
- use user-friendly vars names, but try to find a compromise between the length of the variable name and its clarity for the developer.
- don't use short names for function receiver name
- don't use "failed" word on wrapping errors
- use "any" instead "interface{}"
- use http consts for http return codes
- use http consts for http methods names
- try to encapsulate the structure members as much as possible. Only what is used externally should be public.
- functions must accept logrus.FieldLogger logger as input parameter
- use context where possible
- don't use multi expressions in "if" and other statements. Split to lines
- use lowercase first letters in logger messages
- try split such expression to separate lines:
  "err := downloader.cache.Put(ctx, file.SHA256, file.Dest); err != nil" must be
  "err := downloader.cache.Put(ctx, file.SHA256, file.Dest)
    if err != nil: ... "

**Error handling:**

- use `errors.Is()` and `errors.As()` for error comparison instead of direct comparison
- prefer `fmt.Errorf("...: %w", err)` for error wrapping to preserve the error chain

**Naming & structure:**

- use singular names for packages (e.g., `handler` not `handlers`)
- prefix interface names with behavior verb (e.g., `Reader`, `Writer`, `Validator`)
- use `New<Type>()` constructor pattern for structs that need initialization

**Code organization:**

- group imports: stdlib, external, internal (separated by blank lines)
- keep functions short - if over 40-50 lines, consider splitting
- place exported functions before unexported ones in a file

**Concurrency:**

- always pass `context.Context` as the first parameter
- prefer channels for communication, mutexes for state protection
- use `defer` for cleanup (mutex unlock, file close, etc.)

**Testing:**

- use table-driven tests with descriptive test case names
- use `t.Helper()` in test helper functions
- prefer `testify/assert` or `testify/require` for assertions

**Performance:**

- preallocate slices with `make([]T, 0, capacity)` when size is known
- use `strings.Builder` for string concatenation in loops
