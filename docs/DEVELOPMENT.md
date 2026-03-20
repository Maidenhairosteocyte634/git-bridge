# Development

Guide for building, testing, and contributing to git-bridge.

<br/>

## Table of Contents

- [Prerequisites](#prerequisites)
- [Project Structure](#project-structure)
- [Build](#build)
- [Testing](#testing)
- [Docker](#docker)
- [Workflow](#workflow)
- [CI/CD Workflows](#cicd-workflows)
- [Conventions](#conventions)

<br/>

## Prerequisites

- Go 1.26+
- Make
- Docker (for container builds)

<br/>

## Project Structure

```
.
├── cmd/
│   └── git-bridge/
│       └── main.go                   # Entry point
├── internal/
│   ├── config/                       # Configuration management
│   ├── consumer/                     # Event consumption (SQS, webhooks)
│   ├── mirror/                       # Git mirroring logic
│   ├── notify/                       # Notification handlers (Slack)
│   ├── provider/                     # Git providers (GitHub, GitLab, CodeCommit)
│   └── server/                       # HTTP server (health, webhooks)
├── examples/                         # Configuration examples
├── k8s/                              # Kubernetes manifests
│   ├── namespace.yaml
│   ├── secret.yaml
│   ├── configmap.yaml
│   └── deployment.yaml
├── docs/                             # Documentation
│   ├── ADVANCE.md
│   ├── API.md
│   ├── github-webhook-setup.md
│   ├── gitlab-webhook-setup.md
│   ├── naming-convention.md
│   └── slack-app-setup.md
├── .github/workflows/                # CI/CD pipelines
├── Dockerfile                        # Multi-stage build (alpine)
└── Makefile                          # Build, test, deploy
```

<br/>

### Key Directories

| Directory | Description |
|-----------|-------------|
| `cmd/git-bridge/` | Application entry point |
| `internal/config/` | YAML config loading and validation |
| `internal/consumer/` | SQS long-polling and webhook event consumers |
| `internal/mirror/` | Core git mirroring logic |
| `internal/notify/` | Slack notification handlers |
| `internal/provider/` | GitHub, GitLab, CodeCommit provider implementations |
| `internal/server/` | HTTP server for health checks and webhook endpoints |
| `k8s/` | Kubernetes deployment manifests |

<br/>

## Build

```bash
make build               # Build binary → bin/git-bridge
make run                  # Build and run with examples/config.yaml
make clean                # Remove build artifacts
make tidy                 # Run go mod tidy
```

<br/>

## Testing

```bash
make test                 # Run tests with race detection
make test-cover           # Run tests with coverage report (→ coverage.html)
make fmt                  # Format code
make vet                  # Run go vet
make lint                 # Run golangci-lint
```

<br/>

## Docker

```bash
make docker-build                    # Build Docker image
make docker-push                     # Push image
make docker-buildx-tag               # Multi-arch build (arm64, amd64) with version tag
make docker-buildx-latest            # Multi-arch build with latest tag
make docker-buildx                   # Both version + latest
```

- **Builder**: golang:1.26-alpine
- **Runtime**: alpine:3.23 with git, ca-certificates
- **Port**: 8080

<br/>

## Kubernetes Deployment

```bash
make deploy               # Apply all K8s manifests
make restart              # Restart deployment
make logs                 # Tail pod logs
```

<br/>

## Workflow

```bash
make check-gh                                    # Verify gh CLI is installed and authenticated
make branch name=gitlab-provider                 # Create feature branch from main
make pr title="feat: add GitLab provider"        # Test → push → create PR
```

<br/>

## CI/CD Workflows

| Workflow | Trigger | Description |
|----------|---------|-------------|
| `test.yml` | push, PR, dispatch | Run tests with Go version from go.mod |
| `release.yml` | tag push `v*` | GitHub release with git-cliff changelog |
| `changelog-generator.yml` | PR merge, issue close | Auto-generate CHANGELOG.md |
| `contributors.yml` | after changelog | Auto-generate CONTRIBUTORS.md |
| `gitlab-mirror.yml` | push to main | Mirror to GitLab |
| `stale-issues.yml` | daily cron | Auto-close stale issues |
| `dependabot-auto-merge.yml` | PR (dependabot) | Auto-merge minor/patch updates |
| `issue-greeting.yml` | issue opened | Welcome message |

<br/>

## Conventions

- **Commits**: Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `ci:`, `chore:`)
- **Providers**: GitHub, GitLab, CodeCommit (bidirectional mirroring)
- **Events**: SQS (CodeCommit), Webhooks (GitHub/GitLab)
- **Docker**: Multi-stage alpine, multi-arch support
