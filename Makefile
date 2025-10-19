.PHONY: help test lint build clean install-tools check-coverage security

help: ## Show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

test: ## Run tests
	@echo "🧪 Running tests..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "✅ Tests completed"

test-coverage: test ## Run tests and show coverage
	@echo "📊 Generating coverage report..."
	@go tool cover -html=coverage.out

check-coverage: test ## Check if coverage meets threshold
	@echo "📊 Checking coverage..."
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	threshold=70; \
	echo "Coverage: $${coverage}%"; \
	echo "Threshold: $${threshold}%"; \
	if [ $$(echo "$${coverage} < $${threshold}" | bc -l) -eq 1 ]; then \
		echo "❌ Coverage $${coverage}% is below threshold $${threshold}%"; \
		exit 1; \
	else \
		echo "✅ Coverage $${coverage}% meets threshold $${threshold}%"; \
	fi

lint: ## Run linter
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "✅ Linting completed"; \
	else \
		echo "⚠️  golangci-lint not found. Install it with:"; \
		echo "   brew install golangci-lint"; \
		echo "   or: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

lint-fix: ## Run linter and fix issues
	@echo "🔧 Running linter with auto-fix..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix; \
		echo "✅ Auto-fix completed"; \
	else \
		echo "⚠️  golangci-lint not found. Install it with:"; \
		echo "   brew install golangci-lint"; \
		echo "   or: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

security: ## Run security checks
	@echo "🔒 Running security checks..."
	@govulncheck ./...
	@gosec -quiet ./...
	@echo "✅ Security checks completed"

build: ## Build all packages
	@echo "🏗️  Building packages..."
	@go build -v ./...
	@cd examples/basic && go build -v
	@cd examples/advanced && go build -v
	@echo "✅ Build completed"

clean: ## Clean build artifacts and caches
	@echo "🧹 Cleaning..."
	@go clean -cache -testcache -modcache
	@rm -f coverage.out
	@echo "✅ Cleaned"

fmt: ## Format code
	@echo "💅 Formatting code..."
	@if command -v gofmt >/dev/null 2>&1; then \
		find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./.git/*" | while read -r file; do \
			gofmt -s -w "$$file"; \
		done; \
		echo "✅ Formatting completed"; \
	else \
		echo "❌ gofmt not found. Please install Go: https://go.dev/dl/"; \
		exit 1; \
	fi

vet: ## Run go vet
	@echo "🔎 Running go vet..."
	@go vet ./...
	@echo "✅ Vet completed"

mod: ## Tidy and verify modules
	@echo "📦 Managing modules..."
	@go mod tidy
	@go mod verify
	@echo "✅ Modules updated"

install-tools: ## Install development tools
	@echo "🔧 Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "✅ Tools installed"

run-basic: ## Run basic example
	@echo "🚀 Running basic example..."
	@cd examples/basic && go run main.go

run-advanced: ## Run advanced example
	@echo "🚀 Running advanced example..."
	@cd examples/advanced && go run main.go

# Pre-commit hook - all checks
pre-commit: fmt vet lint test check-coverage security build ## Run all checks before commit
	@echo ""
	@echo "✅ All pre-commit checks passed!"
	@echo "🚀 Ready to commit"

# CI/CD simulation
ci: fmt vet lint test check-coverage security build ## Simulate CI pipeline locally
	@echo ""
	@echo "========================================"
	@echo "✅ CI Pipeline Simulation Complete!"
	@echo "========================================"
	@echo ""
	@echo "All quality gates passed:"
	@echo "  ✅ Code formatting"
	@echo "  ✅ Go vet"
	@echo "  ✅ Linting"
	@echo "  ✅ Tests with race detection"
	@echo "  ✅ Coverage threshold (70%)"
	@echo "  ✅ Security checks"
	@echo "  ✅ Build verification"
	@echo ""
	@echo "🚀 Ready to push!"

all: ci ## Run complete CI pipeline
