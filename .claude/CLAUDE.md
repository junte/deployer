# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Deployer is a secure CI/CD deployment tool that runs as an HTTP/HTTPS service on target servers. CI/CD pipelines (GitLab, GitHub Actions, etc.) trigger deployments via HTTP POST requests — eliminating the need to store private SSH keys in CI/CD providers. The server controls which commands can be executed and validates optional security keys.

## Commands

```bash
make build          # Build native binary to bin/deployer
make build-linux    # Cross-compile for Linux amd64 with version injection
make test           # Run all tests: go test -v ./...
make lint           # Run golangci-lint
make tag            # Create git tag from VERSION file
```

Run a single test file:

```bash
go test -v ./src/core/deployer/...
```

## Architecture

### Server mode (default)

```
HTTP POST Request
    ↓
src/server/main.go        # Parses form data (component, key, async, args)
    ↓
src/core/deployer/main.go # Validates component config and security key
    ↓
src/core/deployer/deployer.go  # Executes shell command via os/exec;
                                # injects request params via Go templates
                                # ({{ .Args.param }}) into command array;
                                # streams stdout/stderr line-by-line
    ↓
src/core/notify/           # Post-deployment Slack notifications
```

### Client mode (`deployer client`)

```
deployer client --url <url> --component <name> [--key <key>] [--arg k=v]...
    ↓
src/client/main.go         # Builds HTTP POST, streams SSE response to stdout,
                            # exits with the remote process exit code
```

```bash
deployer client \
  --url http://localhost:7778 \
  --component app \
  --key 242134321432143214213 \
  --arg tag=v1.2.3
```

**Sync mode**: streams output as SSE (`text/event-stream`) in real-time.
**Async mode** (`async=true`): returns HTTP 200 immediately; deployment runs in background goroutine.

## Configuration

Config is loaded from `config.yaml` (Viper). See `config.yaml.example` for the full schema. Key fields:

```yaml
port: ":7777"
environment: "dev"
tls:                      # optional HTTPS
  cert: ./tls/cert.crt
  key: ./tls/cert.key
notification:             # optional global Slack
  slack:
    apiToken: "<token>"
    channel: "#deployments"
components:
  backend:
    command: ["/opt/deploy.sh", "{{ .Args.tag }}"]  # Go template args
    key: "<secret>"         # optional auth key
    workdir: "/opt/app"     # optional working dir
    notification:           # optional component-level Slack override
      slack:
        channel: "#backend-deploys"
```

Development config lives in `tools/dev/config.yaml`. Example HTTP requests are in `http/api.http`.

## Key Types

- `src/core/types.go` — `ComponentDeployRequest` (input) and `ComponentDeployResults` (output with stdout/stderr/exit code)
- `src/config/main.go` — `AppConfig`, `ComponentConfig`, `TLSConfig`, `NotificationConfig`
- `src/client/main.go` — `Options` (client request config)

## Module

Module path: `deployer` (Go 1.26). Binary entry point: `src/main.go`.

## Development Workflow

Follow this workflow when writing or modifying Go code:

1. **Implement** — write or modify the Go code following `go-codestyle.md` rules
2. **Write tests** — add or update tests for changed logic; use table-driven tests
3. **Run tests** — `make test` (or `go test -v ./path/to/pkg/...` for a single package); fix any failures before continuing
4. **Lint** — use `lint-runner` subagent; it runs `make lint` and fixes all reported issues
5. **Code review** — use `go-reviewer` subagent to check against codestyle rules
6. **Security audit** — use `security-auditor` subagent when auth, key handling, command execution, or config parsing is touched

Never skip steps. Always fix issues before moving to the next step.

## Subagents

Use these subagents automatically when the situation matches — no need to ask.

| Agent | When to use |
|---|---|
| `go-reviewer` | After writing or modifying any Go file — review against codestyle rules |
| `security-auditor` | When auth, key handling, command execution, or config parsing is touched |
| `test-runner` | When asked to run tests or when a test failure needs to be diagnosed and fixed |
| `lint-runner` | After writing or modifying any Go file — run `make lint` and fix any reported issues |

## Notes

- Always use Context7 MCP when I need library/API documentation, code generation, setup or configuration steps without me having to explicitly ask
