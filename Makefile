# Cross-platform compatibility
ifeq ($(OS),Windows_NT)
	BINARY_EXT := .exe
	RM_CMD := if exist coverage.out del coverage.out & if exist coverage.html del coverage.html
	GOBIN := $(USERPROFILE)\go\bin
else
	BINARY_EXT :=
	RM_CMD := rm -f coverage.out coverage.html
	GOBIN := $(HOME)/go/bin
endif

.PHONY: test tdd test-watch test-coverage fmt fmt-check lint lint-fix typos typos-fix deps clean ci help

# Test targets
test:
	go test ./...

test-verbose:
	$(GOBIN)/gotestsum$(BINARY_EXT) --format testname ./...

tdd:
	$(GOBIN)/gotestsum$(BINARY_EXT) --watch --format testname ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format targets
fmt:
	go fmt ./...

fmt-check:
ifeq ($(OS),Windows_NT)
	@for /f %%i in ('gofmt -l .') do @echo Code needs formatting: %%i && exit 1
else
	@test -z "$$(gofmt -l .)" || (echo "Code needs formatting. Run 'make fmt'" && exit 1)
endif

# Lint targets
lint:
	$(GOBIN)/golangci-lint$(BINARY_EXT) run ./...

lint-fix:
	$(GOBIN)/golangci-lint$(BINARY_EXT) run --fix ./...

# Spell check targets
typos:
	@command -v typos >/dev/null 2>&1 && typos || echo "typos not installed, skipping"

typos-fix:
	typos --write-changes

# Development targets
deps:
	go install gotest.tools/gotestsum@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@echo ""
	@echo "Note: typos must be installed separately:"
	@echo "  macOS:   brew install typos-cli"
	@echo "  Linux:   cargo install typos-cli"

clean:
	@$(RM_CMD)

# CI target
ci: fmt-check test lint typos

# Help target
help:
	@echo "Claude Code SDK for Go - Development Commands"
	@echo ""
	@echo "Test Commands:"
	@echo "  test           Run all tests"
	@echo "  test-verbose   Run tests with detailed output"
	@echo "  tdd            Run tests in watch mode"
	@echo "  test-coverage  Run tests with coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt            Format all Go code"
	@echo "  fmt-check      Check if code is formatted (CI)"
	@echo "  lint           Run linter on all code"
	@echo "  lint-fix       Run linter and auto-fix issues"
	@echo "  typos          Check for spelling errors"
	@echo "  typos-fix      Auto-fix spelling errors"
	@echo ""
	@echo "Utilities:"
	@echo "  clean          Remove build artifacts"
	@echo "  deps           Install development tools"
	@echo "  ci             Run full CI pipeline locally"
	@echo "  help           Show this help message"
