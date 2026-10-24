# generated_by: forge:harness — do not edit inside the managed block; edit meta/.skills.yml or run /forge:harness
# >>> forge:harness (managed) >>>
.DEFAULT_GOAL := help
PKGS ?= ./...
TEST_WORKERS ?= $(shell n=$$( (nproc || sysctl -n hw.ncpu) 2>/dev/null ); echo $$(( n > 1 ? n / 2 : 1 )) )
TESTPKGS ?=
AUDIT_DIR ?= $(CURDIR)/.audit

.PHONY: help build run test test-race test-impact coverage audit test-e2e test-all lint fmt tidy ci clean

help: ## list targets
	@grep -hE '^[a-zA-Z0-9_-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-14s %s\n", $$1, $$2}'

build: ## compile everything
	go build $(PKGS)

run: ## run the board locally (PORT, default 8080; ADR-0003: bare binary on the host)
	go run ./cmd/board

test: ## fast unit suite (browser e2e tests TestE2E_* are excluded; see test-e2e)
	go test -p $(TEST_WORKERS) -skip '^TestE2E_' $(PKGS)

test-race: ## unit suite under the race detector (the concurrency guarantees of ADR-0006)
	go test -race -p $(TEST_WORKERS) -skip '^TestE2E_' $(PKGS)

test-impact: ## unit tests for TESTPKGS + packages importing them (empty -> make test)
ifeq ($(strip $(TESTPKGS)),)
	$(MAKE) test
else
	go build ./...
	@changed=$$(go list $(TESTPKGS)); \
	impacted=$$(go list -f '{{.ImportPath}} {{join .Deps " "}} {{join .TestImports " "}} {{join .XTestImports " "}}' ./... \
	  | grep -Fw "$$changed" | cut -d' ' -f1); \
	go test -p $(TEST_WORKERS) -skip '^TestE2E_' $$(echo "$$changed $$impacted" | tr ' ' '\n' | sort -u)
endif

coverage: ## coverage over COVERAGE_FILES (kanban's review.coverage_min floor)
	@prof="$$(mktemp)"; trap 'rm -f "$$prof"' EXIT; \
	go test -skip '^TestE2E_' -coverprofile="$$prof" ./... && \
	go tool cover -func="$$prof" | awk -v files="$$COVERAGE_FILES" \
	  'BEGIN { n = split(files, f, " ") } $$1 != "total:" { \
	    p = $$1; sub(/:[0-9]+:.*$$/, "", p); keep = (n == 0); \
	    for (i = 1; i <= n; i++) if (p == f[i] || (length(p) > length(f[i]) \
	      && substr(p, length(p) - length(f[i])) == "/" f[i])) keep = 1; \
	    if (keep) { sub("%", "", $$3); s += $$3; c++ } } \
	  END { if (n > 0 && c == 0) print "coverage: n/a"; \
	    else printf "coverage: %.1f%%\n", (c ? s / c : 0) }'

audit: ## dependency scan + licence report into AUDIT_DIR (kanban done step 2c)
	@mkdir -p "$(AUDIT_DIR)"; d="$(AUDIT_DIR)"; \
	if ! command -v govulncheck >/dev/null 2>&1; then echo not-installed > "$$d/govulncheck.status"; \
	elif govulncheck -json ./... > "$$d/govulncheck.json" 2> "$$d/govulncheck.err"; then echo ok > "$$d/govulncheck.status"; \
	elif grep -qiE 'dial tcp|no such host|connection refused|i/o timeout|network is unreachable|TLS handshake' "$$d/govulncheck.err"; \
	then echo scanner-unavailable > "$$d/govulncheck.status"; \
	else echo failed > "$$d/govulncheck.status"; fi; \
	if ! command -v go-licenses >/dev/null 2>&1; then echo not-installed > "$$d/licenses.status"; \
	elif go-licenses report ./... > "$$d/licenses.raw" 2> "$$d/licenses.err"; then \
	  awk -F, 'NF>=3 {print $$1","$$NF}' "$$d/licenses.raw" > "$$d/licenses.csv" && echo ok > "$$d/licenses.status"; \
	else echo failed > "$$d/licenses.status"; fi

test-e2e: ## browser e2e: Go chromedp tests TestE2E_* (ADR-0010); needs Chrome or Chromium; fails if there are none
	@go test -list '^TestE2E_' ./... 2>/dev/null | grep -q '^TestE2E_' || { echo "test-e2e: no TestE2E_ tests found (ADR-0010)" >&2; exit 1; }
	go test -count=1 -p 1 -run '^TestE2E_' ./...

test-all: test test-e2e ## unit + browser e2e

lint: ## static checks
	@command -v golangci-lint >/dev/null || { echo "lint: tool-missing golangci-lint" >&2; exit 127; }
	go vet $(PKGS) && golangci-lint run

fmt: ## format
	gofmt -l -w .

tidy: ## tidy go.mod
	go mod tidy

ci: lint build test ## local gate: lint, build, test

clean: ## remove build artifacts
	rm -rf bin coverage.out .audit
# <<< forge:harness (managed) <<<
