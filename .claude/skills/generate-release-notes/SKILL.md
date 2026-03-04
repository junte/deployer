---
name: generate-release-notes
description: Generate GitHub release notes in markdown from git commits since a given tag. Use this skill whenever the user asks to "generate release notes", "create changelog", "what changed since v1.2.3", "prepare release notes for v2.0", or wants a summary of changes between tags. Trigger even if the user just says "release notes" without specifying a tag — auto-detect the latest tag in that case.
---

# Release Notes Generator

Generate clean GitHub release notes in markdown by reading git commits since a specified tag (or the latest tag if none is given).

## Step 1: Determine the base tag

If the user provided a tag (e.g., `v1.2.3`), use it directly.

If no tag was provided, find the latest tag:

```bash
git tag --sort=-version:refname | head -1
```

Confirm the tag you'll use before continuing (one line: "Generating release notes since `<tag>`…").

## Step 2: Collect commits since the tag

```bash
git log <tag>..HEAD --pretty=format:"%s"
```

If `git log` returns nothing, report "No commits since `<tag>`." and stop.

## Step 3: Categorize commits

Parse each commit subject using conventional commit prefixes:

| Prefix | Section |
|---|---|
| `feat:` / `feat(…):` | ✨ What's New |
| `fix:` / `fix(…):` | 🐛 Bug Fixes |
| Any commit with `BREAKING CHANGE:` in body, or `!` after type (e.g., `feat!:`) | ⚠️ Breaking Changes |
| `perf:` | ⚡ Performance |
| `refactor:`, `style:`, `docs:`, `build:`, `ci:`, `chore:`, `test:` | 🔧 Other Changes |

Rules:

- Strip the prefix (`feat:`, `fix(scope):`, etc.) — keep only the subject text.
- Capitalize the first letter of each entry.
- If a commit belongs to Breaking Changes, also include it in its original section (feat/fix/etc.) — duplication is intentional.
- Omit sections that have no entries.
- Omit `chore:`, `ci:`, `test:`, `style:`, and `docs:` commits from the **Other Changes** section unless there are fewer than 3 total commits (in that case include everything to avoid an empty release).

## Step 4: Output the release notes

Use this exact template:

```markdown
## What's Changed

### ✨ What's New
- <entry>
- <entry>

### 🐛 Bug Fixes
- <entry>

### ⚠️ Breaking Changes
- <entry>

### ⚡ Performance
- <entry>

### 🔧 Other Changes
- <entry>

**Full Changelog**: https://github.com/junte/deployer/compare/<tag>...HEAD
```

Only include sections that have at least one entry. The "Full Changelog" line is always present.

If the user asks for a specific next version (e.g., "for v1.3.0"), replace `HEAD` in the Full Changelog URL with that version tag.

Output the markdown in a code block so it's easy to copy.
