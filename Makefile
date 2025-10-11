GO        ?= go
PKG       ?= ./...
TAGS      ?=
FUZZTIME  ?= 20s
COVERFILE ?= coverage.out
COVERHTML ?= coverage.html

GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null)
STATICCHECK   := $(shell command -v staticcheck 2>/dev/null)
GOSEC         := $(shell command -v gosec 2>/dev/null)
GOVULNCHECK   := $(shell command -v govulncheck 2>/dev/null)
GOIMPORTS     := $(shell command -v goimports 2>/dev/null)

TEST_FLAGS   ?=
BENCH_FLAGS  ?= -run=^$$ -bench=. -benchmem
RACE_FLAG    ?= -race

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[1mTargets\033[0m\n"} \
	/^[a-zA-Z0-9_.-]+:.*##/ { printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: fmt
fmt: ## Run go fmt
	@$(GO) fmt $(PKG)
	@if [ -n "$(GOIMPORTS)" ]; then \
		echo "[goimports] fixing imports"; \
		$(GOIMPORTS) -w . ; \
	else \
		echo "[skip] goimports not found (install with: go install golang.org/x/tools/cmd/goimports@latest)"; \
	fi

.PHONY: vet
vet: ## Run go vet
	@$(GO) vet $(PKG)

.PHONY: lint
lint: ## Run linters
	@if [ -n "$(GOLANGCI_LINT)" ]; then \
		echo "[golangci-lint] running"; \
		$(GOLANGCI_LINT) run; \
	else \
		echo "[skip] golangci-lint not found (install with: make tools.lint)"; \
	fi
	@if [ -n "$(STATICCHECK)" ]; then \
		echo "[staticcheck] running"; \
		$(STATICCHECK) ./... ; \
	else \
		echo "[skip] staticcheck not found (install with: make tools.staticcheck)"; \
	fi

.PHONY: sec
sec: ## Run gosec
	@if [ -n "$(GOSEC)" ]; then \
		echo "[gosec] scanning ./..."; \
		$(GOSEC) ./... ; \
	else \
		echo "[skip] gosec not found (install with: make tools.gosec)"; \
	fi

.PHONY: vuln
vuln: ## Run govulncheck
	@if [ -n "$(GOVULNCHECK)" ]; then \
		echo "[govulncheck] scanning module"; \
		$(GOVULNCHECK) ./... ; \
	else \
		echo "[skip] govulncheck not found (install with: make tools.govulncheck)"; \
	fi

.PHONY: test
test: ## Run unit tests
	@$(GO) test $(TEST_FLAGS) -tags '$(TAGS)' $(PKG)

.PHONY: test.race
test.race: ## Run unit tests with race detector
	@$(GO) test $(TEST_FLAGS) $(RACE_FLAG) -tags '$(TAGS)' $(PKG)

.PHONY: cover
cover: ## Run tests with coverage, print summary, and write HTML report
	@$(GO) test -coverprofile=$(COVERFILE) -tags '$(TAGS)' $(PKG)
	@$(GO) tool cover -func=$(COVERFILE)
	@$(GO) tool cover -html=$(COVERFILE) -o $(COVERHTML)
	@echo "Coverage HTML: $(COVERHTML)"

.PHONY: bench
bench: ## Run benchmarks
	@$(GO) test $(BENCH_FLAGS) -tags '$(TAGS)' $(PKG)

.PHONY: tidy
tidy: ## Go module tidy
	@$(GO) mod tidy

.PHONY: clean
clean: ## Clean build/test artifacts
	@go clean -testcache
	@rm -f $(COVERFILE) $(COVERHTML) 2>/dev/null || true

.PHONY: ci
ci: fmt vet lint sec vuln test ## CI: format, vet, lint, security, unit tests

.PHONY: tools
tools: tools.lint tools.staticcheck tools.gosec tools.govulncheck tools.goimports

.PHONY: tools.lint
tools.lint: ## Install golangci-lint
	@echo "Installing golangci-lint..."
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Done. Ensure $$GOBIN or \$$(go env GOPATH)/bin is in your PATH."

.PHONY: tools.staticcheck
tools.staticcheck: ## Install staticcheck
	@echo "Installing staticcheck..."
	@$(GO) install honnef.co/go/tools/cmd/staticcheck@latest
	@echo "Done."

.PHONY: tools.gosec
tools.gosec: ## Install gosec
	@echo "Installing gosec..."
	@$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "Done."

.PHONY: tools.govulncheck
tools.govulncheck: ## Install govulncheck
	@echo "Installing govulncheck..."
	@$(GO) install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "Done."

.PHONY: tools.goimports
tools.goimports: ## Install goimports
	@echo "Installing goimports..."
	@$(GO) install golang.org/x/tools/cmd/goimports@latest
	@echo "Done."
