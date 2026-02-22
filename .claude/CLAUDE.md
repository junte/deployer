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
go test -v ./internal/core/deployer/...
```

## Architecture

```
HTTP POST Request
    ↓
internal/server/main.go        # Parses form data (component, key, async, args)
    ↓
internal/core/deployer/main.go # Validates component config and security key
    ↓
internal/core/deployer/deployer.go  # Executes shell command via os/exec;
                                    # injects request params via Go templates
                                    # ({{ .Args.param }}) into command array;
                                    # streams stdout/stderr line-by-line
    ↓
internal/core/notify/           # Post-deployment Slack notifications
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

Development config lives in `dev/config.yaml`. Example HTTP requests are in `http/api.http`.

## Key Types

- `internal/core/types.go` — `ComponentDeployRequest` (input) and `ComponentDeployResults` (output with stdout/stderr/exit code)
- `internal/config/main.go` — `AppConfig`, `ComponentConfig`, `TLSConfig`, `NotificationConfig`

## Module

Module path: `deployer` (Go 1.19). Binary entry point: `cmd/server/main.go`.

## Subagents

Use these subagents automatically when the situation matches — no need to ask.

| Agent | When to use |
|---|---|
| `go-reviewer` | After writing or modifying any Go file — review against codestyle rules |
| `security-auditor` | When auth, key handling, command execution, or config parsing is touched |
| `test-runner` | When asked to run tests or when a test failure needs to be diagnosed and fixed |

## Notes

- Always use Context7 MCP when I need library/API documentation, code generation, setup or configuration steps without me having to explicitly ask
