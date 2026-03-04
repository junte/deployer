# Architecture Guidelines

Soft guidelines for structuring code. Use "prefer" thinking — these are defaults, not absolutes.

## File Size

- Aim for ~200-300 lines per file as a soft ceiling
- A file growing beyond this is a signal to check whether it holds multiple responsibilities
- Size alone is not a reason to split — cohesion matters more than line count

## Single Responsibility

- One concern per file, one purpose per package
- A file should be describable in one short sentence (e.g., "parses deployment config", "handles HTTP routing")
- If you need "and" to describe what a file does, consider splitting

## Package Cohesion

- Group by domain, not by technical role
- Prefer `deployer/`, `notify/`, `config/` over `models/`, `helpers/`, `utils/`
- A package should represent a concept, not a grab-bag of similar-looking code

## Layer Separation

- Don't mix transport (HTTP handlers), business logic, and infrastructure (exec, filesystem) in the same file
- This project's layers: `server` (transport) → `core` (business logic) → `config` (configuration)
- Keep each layer focused — a handler should delegate to core, not implement business rules inline

## Dependency Direction

- Depend inward: transport → core → types
- Never import a transport package from core, or core from types
- If two packages need each other, extract the shared concept into a third package or use interfaces

## Interface Boundaries

- Define interfaces where they are consumed, not where they are implemented
- A small, focused interface (1-3 methods) is easier to mock and test
- Don't create interfaces preemptively — extract them when a second implementation or a test mock appears

## When to Split

Consider splitting a file or package when:

- It contains multiple unrelated types or concepts
- It has a long import list pulling from many different domains
- It's hard to name without using "and" or generic words like "utils"
- Different parts change for different reasons

## When NOT to Split

Avoid splitting when:

- The code is only used in one place — a single-use package adds indirection without value
- Splitting would create packages with only one small file
- The pieces are tightly coupled and would need to import each other
- You're splitting for predicted future needs that don't exist yet

## Flat Over Nested

- Prefer shallow directory trees — one or two levels under `src/` is usually enough
- Deep nesting (`src/core/deploy/internal/executor/runner/`) makes navigation harder
- A flat list of well-named packages is easier to reason about than a deep hierarchy

## New File/Package Checklist

Before creating a new file or package, consider:

1. Can this live in an existing file without hurting clarity?
2. Does the new package have a clear, singular name that describes its purpose?
3. Will it have more than one consumer, or is it only used in one place?
4. Does it introduce a circular dependency?
5. Would a future reader find it where they'd expect to look?

If the answer to #1 is yes, prefer adding to the existing file.
