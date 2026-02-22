# Commit Message Guidelines

Follow these rules when creating commit messages for this project.

## Format

```text
<type>: <subject>

[optional body]

[optional footer]
```

## Types

- **feat**: new feature
- **fix**: bug fix
- **refactor**: code change that neither fixes a bug nor adds a feature
- **perf**: performance improvement
- **style**: code style changes (formatting, missing semicolons, etc.)
- **test**: adding or updating tests
- **docs**: documentation changes
- **build**: changes to build system or dependencies
- **ci**: CI/CD configuration changes
- **chore**: other changes that don't modify src or test files

## Subject Line Rules

- use imperative mood ("add feature" not "added feature" or "adds feature")
- don't capitalize first letter
- no period at the end
- limit to 50 characters
- be specific and concise

## Body Rules

- separate from subject with blank line
- wrap at 72 characters
- explain *what* and *why*, not *how*
- use imperative mood
- can include multiple paragraphs

## Examples

### Simple commit

```text
fix: handle nil pointer in download manager
```

### With body

```text
feat: add retry mechanism for failed downloads

implement exponential backoff strategy for retrying failed downloads.
this improves reliability when dealing with unstable network connections.
```

### Breaking change

```text
refactor: change config structure

BREAKING CHANGE: config format has changed from YAML to TOML.
users need to migrate their config files.
```

### Multiple changes

```text
feat: add concurrent downloads

- implement worker pool for parallel downloads
- add configuration for max concurrent downloads
- include download progress tracking
```

## Common Mistakes to Avoid

- ❌ "Fixed bug" → ✅ "fix: handle empty response body"
- ❌ "feat: Added new feature." → ✅ "feat: add resume capability"
- ❌ "updated code" → ✅ "refactor: simplify error handling"
- ❌ "Fix: Bug fix" → ✅ "fix: prevent race condition in worker pool"
