.PHONY: help all clean lint test build bench examples e2e-build e2e-test \
        e2e-one e2e-mod-audit e2e-mod-clean e2e-mod-list e2e-mod-get e2e-mod-sum e2e-mod-update \
        test-hotspots coverage-short zip

# Print Makefile target help by scanning for lines with '##' descriptions.
help: ## Show this help with targets and descriptions
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "%-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Benchmark configuration (override via: make bench BENCH=... BENCHTIME=...)
BENCH ?= BenchmarkAMI_Subcommands
BENCHTIME ?= 1x

all: build ## Default: build the ami CLI

clean: ## Remove and recreate the build/ directory
	rm -rf ./build
	mkdir -p ./build

lint: check-single-test check-single-declaration test-hotspots
	## Run static checks: go vet + single-test-per-file + single-declaration-per-file
	# Enforce one Test* per *_test.go (repo-wide)
	go vet -v ./...

check-single-test: ## Enforce one Test* per *_test.go across src/
	go run ./scripts/check-single-test-per-file.go src/

.PHONY: check-single-declaration
check-single-declaration: ## Enforce single cohesive declaration per .go file across src/
	CHECK_TEST_MODE=package bash ./scripts/check-single-declaration-per-file.sh src/

test: lint  ## Run all tests (go test -v ./...)
	go test -v ./...

coverage-short: ## Fast coverage on CLI (filters heavy tests) + sanity on schemas
	@echo "Running short coverage for CLI (filtered) ..."
	mkdir -p build
	# Run only fast CLI tests: Root/Version/Help/Mod/Lint/Pipeline; skip Build/Test E2E-like suites
	GOFLAGS= go test -count=1 -short -timeout 90s \
	  -run '^(TestRoot_|TestVersion_|TestHelp_|TestMod_|TestLint_|TestPipeline_)' \
	  -covermode=atomic -coverprofile=build/coverage-short.out ./src/
	@echo "CLI coverage written to build/coverage-short.out"
	# Sanity run schema packages in short mode (no coverage merge)
	GOFLAGS= go test -count=1 -short ./src/schemas/log ./src/

build: clean ## Build the ami CLI binary to build/ami
	go build -o build/ami ./src/cmd/ami

zip: ## Create a repository snapshot zip at build/repo.zip (tracked files only)
	mkdir -p build
	git archive -o build/repo.zip HEAD

# Run CLI microbenchmarks for ami subcommands in isolated sandboxes.
bench: ## Run CLI microbenchmarks (vars: BENCH, BENCHTIME)
	@echo "Running CLI benchmarks: $(BENCH) (benchtime=$(BENCHTIME))"
	go test -run ^$$ -bench $(BENCH) -benchtime=$(BENCHTIME) ./src/cmd/ami

gen-diag-codes: ## Generate docs/diag-codes.md from code annotations
	@echo "Generating docs/diag-codes.md ..."
	go run ./tools/gen-diag-codes

e2e-build: ## Build CLI for end-to-end tests
	@echo "Building ami CLI for E2E..."
	go build -o build/ami ./src/cmd/ami

e2e-test: e2e-build ## Run all E2E CLI tests (tests/e2e)
	@echo "Running E2E CLI tests (tests/e2e)..."
	go test -v ./tests/e2e

# Run a subset of E2E tests by regex name. Usage:
#   make e2e-one NAME=AmiModGet
e2e-one: e2e-build ## Run E2E tests matching NAME=Pattern (e.g., NAME=AmiModGet)
	@if [ -z "$(NAME)" ]; then \
	  echo "NAME required, e.g., make e2e-one NAME=AmiModGet"; \
	  exit 1; \
	fi
	@echo "Running E2E tests matching: $(NAME)"
	go test -v ./tests/e2e -run "$(NAME)"

# Convenience targets per mod subcommand
e2e-mod-audit: e2e-build ## Run only E2E tests for 'ami mod audit'
	go test -v ./tests/e2e -run AmiModAudit

e2e-mod-clean: e2e-build ## Run only E2E tests for 'ami mod clean'
	go test -v ./tests/e2e -run AmiModClean

e2e-mod-list: e2e-build ## Run only E2E tests for 'ami mod list'
	go test -v ./tests/e2e -run AmiModList

e2e-mod-get: e2e-build ## Run only E2E tests for 'ami mod get'
	go test -v ./tests/e2e -run AmiModGet

e2e-mod-sum: e2e-build ## Run only E2E tests for 'ami mod sum'
	go test -v ./tests/e2e -run AmiModSum

e2e-mod-update: e2e-build ## Run only E2E tests for 'ami mod update'
	go test -v ./tests/e2e -run AmiModUpdate

# List packages and files under src/ missing test coverage patterns.
# - Reports packages with zero *_test.go files.
# - Reports .go files without a matching *_test.go sibling (same basename).
test-hotspots: ## Report packages/files missing test coverage pairs (enforced: packages only)
	@bash ./scripts/test-hotspots.sh

examples: ## Build example workspaces and stage outputs under build/examples/
	# Build all example workspaces and stage their outputs under build/examples/
	rm -rf build/examples
	mkdir -p build/examples
	set -e; \
	for d in examples/*; do \
	  if [ -f "$$d/ami.workspace" ]; then \
	    echo "Building example: $$d"; \
	    name=$$(basename "$$d"); \
	    ( cd "$$d" && { [ -x ../../build/ami ] && ../../build/ami build --verbose || go run ../../src/cmd/ami build --verbose; } ); \
	    mkdir -p "build/examples/$$name"; \
	    if [ -d "$$d/build" ]; then \
	      cp -R "$$d/build/." "build/examples/$$name/"; \
	    fi; \
	  fi; \
	done

.PHONY: example-time-ticker
example-time-ticker: build ## Build only the time-ticker example (emits IR/ASM under examples/time-ticker/build)
	@echo "Building example: examples/time-ticker (with --verbose for IR/ASM)"
	cd examples/time-ticker && ../../build/ami build --verbose || (echo "Falling back to go run" && cd examples/time-ticker && go run ../../src/cmd/ami build --verbose)
	mkdir -p build/examples/time-ticker
	if [ -d examples/time-ticker/build ]; then cp -R examples/time-ticker/build/. build/examples/time-ticker/; fi
