---
name: check
description: Run the full quality pipeline for this Go project — tests, lint, code review, and security audit — in sequence with early stopping. Use this skill whenever the user invokes `/check`, says "run all checks", "validate my changes", "is the code ready?", "run the pipeline", "check before I push", or asks whether their implementation is done. Also use proactively after completing any non-trivial Go code change.
---

# Check — Full Quality Pipeline

You are orchestrating this project's quality pipeline. The goal is to catch all problems before code is pushed: failing tests, lint violations, code style and logic issues, and security vulnerabilities.

Run the steps below **in order**. Steps 1 and 2 are blocking — there is no value in reviewing code that fails tests or has lint errors.

## Step 0: Identify changed files

Before running agents, determine which files changed:

```bash
git diff --name-only HEAD
```

If there are staged-but-not-committed changes: `git diff --name-only --cached`. If the working tree is clean, check the last commit: `git diff --name-only HEAD~1 HEAD`.

Keep this list — you need it to decide whether to run the security audit in Step 4.

## Step 1: Tests (blocking)

Invoke the `test-runner` agent.

- **If tests fail**: report the failure summary and stop. Do not proceed to lint, code review, or security audit.
- **If tests pass**: continue to Step 2.

## Step 2: Lint (blocking)

Invoke the `lint-runner` agent.

- **If lint violations remain after the agent's fix pass**: report the outstanding issues and stop. Do not proceed to code review.
- **If lint passes**: continue to Step 3.

## Step 3: Code Review

Invoke the `go-reviewer` agent. This step always runs when tests and lint have passed.

## Step 4: Security Audit (conditional)

Invoke the `security-auditor` agent **only if** any changed file touches security-sensitive areas:

- Key validation or authentication (`deployer/main.go`, files containing `key`, `auth`, `token`)
- Command execution (`deployer.go`, any file using `os/exec`)
- Config loading (`config/main.go`, `config.yaml`)
- HTTP request handling (`server/main.go`)

If no changed file matches any of those areas, skip this step and note it in the summary.

## Summary

After all applicable steps complete, output this table:

```
## Check Results

| Step           | Status                         |
|----------------|--------------------------------|
| Tests          | ✅ Pass / ❌ Fail              |
| Lint           | ✅ Pass / ❌ Fail / ⏭ Skipped |
| Code Review    | ✅ Done / ⏭ Skipped           |
| Security Audit | ✅ Done / ⏭ Not applicable    |
```

If code review or security audit surfaced **Critical** or **High** severity findings, list them below the table with file path and line references. If everything passed cleanly, say so explicitly.
