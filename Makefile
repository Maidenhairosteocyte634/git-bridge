APP_NAME := git-bridge
BUILD_DIR := bin
IMG ?= somaz940/git-bridge:v0.1.0
CONTAINER_TOOL ?= docker
PLATFORMS ?= linux/arm64,linux/amd64

.PHONY: all build run test clean docker-build docker-push fmt vet lint

all: fmt vet build

## Build
build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/git-bridge

run: build
	CONFIG_PATH=examples/config.yaml $(BUILD_DIR)/$(APP_NAME)

## Test
test:
	@go test -race ./internal/... 2>&1 | grep -E "^(ok|FAIL|\?)" | while read line; do \
		pkg=$$(echo "$$line" | awk '{print $$2}'); \
		status=$$(echo "$$line" | awk '{print $$1}'); \
		printf "  %-6s %s\n" "$$status" "$$pkg"; \
	done

test-cover:
	@go test -race -coverprofile=coverage.out ./internal/... 2>&1 | grep -E "^ok" | while read line; do \
		pkg=$$(echo "$$line" | awk '{print $$2}'); \
		cov=$$(echo "$$line" | grep -o '[0-9]*\.[0-9]*%'); \
		printf "  %-40s %s\n" "$$pkg" "$$cov"; \
	done
	@go tool cover -html=coverage.out -o coverage.html
	@echo "  ----------------------------------------"
	@go tool cover -func=coverage.out | grep "^total:" | awk '{printf "  %-40s %s\n", "total", $$NF}'

## Code Quality
fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

## Docker
docker-build:
	$(CONTAINER_TOOL) build -t ${IMG} .

docker-push:
	$(CONTAINER_TOOL) push ${IMG}

docker-buildx-tag:
	- $(CONTAINER_TOOL) buildx create --name git-bridge-builder
	$(CONTAINER_TOOL) buildx use git-bridge-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) \
		--tag ${IMG} \
		-f Dockerfile .
	- $(CONTAINER_TOOL) buildx rm git-bridge-builder

docker-buildx-latest:
	- $(CONTAINER_TOOL) buildx create --name git-bridge-builder
	$(CONTAINER_TOOL) buildx use git-bridge-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) \
		--tag $(shell echo ${IMG} | cut -f1 -d:):latest \
		-f Dockerfile .
	- $(CONTAINER_TOOL) buildx rm git-bridge-builder

docker-buildx: docker-buildx-tag docker-buildx-latest

## K8s Deploy
deploy:
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/secret.yaml
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/deployment.yaml

restart:
	kubectl rollout restart -n git-bridge deployment/git-bridge

logs:
	kubectl logs -n git-bridge -l app=git-bridge -f

## Cleanup
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

## Dependency
tidy:
	go mod tidy

## Workflow

check-gh: ## Check if gh CLI is installed and authenticated
	@command -v gh >/dev/null 2>&1 || { echo "\033[31m✗ gh CLI not installed. Run: brew install gh\033[0m"; exit 1; }
	@gh auth status >/dev/null 2>&1 || { echo "\033[31m✗ gh CLI not authenticated. Run: gh auth login\033[0m"; exit 1; }
	@echo "\033[32m✓ gh CLI ready\033[0m"

branch: ## Create feature branch (usage: make branch name=watch-mode)
	@if [ -z "$(name)" ]; then echo "Usage: make branch name=<feature-name>"; exit 1; fi
	git checkout main
	git pull origin main
	git checkout -b feat/$(name)
	@echo "\033[32m✓ Branch feat/$(name) created\033[0m"

pr: check-gh ## Run tests, push, and create PR (usage: make pr title="Add feature")
	@if [ -z "$(title)" ]; then echo "Usage: make pr title=\"PR title\""; exit 1; fi
	go test ./... -race -cover
	go vet ./...
	git push -u origin $$(git branch --show-current)
	@./scripts/create-pr.sh "$(title)"
	@echo "\033[32m✓ PR created\033[0m"

## Help
help:
	@echo "Usage:"
	@echo "  make build              - Build binary"
	@echo "  make run                - Build and run locally"
	@echo "  make test               - Run tests"
	@echo "  make test-cover         - Run tests with coverage"
	@echo "  make fmt                - Format code"
	@echo "  make vet                - Run go vet"
	@echo "  make lint               - Run golangci-lint"
	@echo "  make docker-build       - Build Docker image"
	@echo "  make docker-push        - Push Docker image"
	@echo "  make docker-buildx-tag  - Build and push multi-arch image (version tag)"
	@echo "  make docker-buildx-latest - Build and push multi-arch image (latest tag)"
	@echo "  make docker-buildx      - Build and push multi-arch image (both tags)"
	@echo "  make deploy             - Deploy to Kubernetes"
	@echo "  make restart            - Restart deployment"
	@echo "  make logs               - Tail pod logs"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make tidy               - Run go mod tidy"
	@echo "  make check-gh           - Check gh CLI is ready"
	@echo "  make branch name=X      - Create feature branch"
	@echo "  make pr title=\"X\"       - Test, push, create PR"
