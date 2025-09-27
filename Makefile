.PHONY: all clean lint test build examples e2e-build e2e-test \
        e2e-one e2e-mod-audit e2e-mod-clean e2e-mod-list e2e-mod-get e2e-mod-sum e2e-mod-update \
        test-hotspots

all: build

clean:
	rm -rf ./build
	mkdir -p ./build

lint:
	go vet -v ./...

test:
	go test -v ./...

build: clean
	go build -o build/ami ./src/cmd/ami

e2e-build:
	@echo "Building ami CLI for E2E..."
	go build -o build/ami ./src/cmd/ami

e2e-test: e2e-build
	@echo "Running E2E CLI tests (tests/e2e)..."
	go test -v ./tests/e2e

# Run a subset of E2E tests by regex name. Usage:
#   make e2e-one NAME=AmiModGet
e2e-one: e2e-build
	@if [ -z "$(NAME)" ]; then \
	  echo "NAME required, e.g., make e2e-one NAME=AmiModGet"; \
	  exit 1; \
	fi
	@echo "Running E2E tests matching: $(NAME)"
	go test -v ./tests/e2e -run "$(NAME)"

# Convenience targets per mod subcommand
e2e-mod-audit: e2e-build
	go test -v ./tests/e2e -run AmiModAudit

e2e-mod-clean: e2e-build
	go test -v ./tests/e2e -run AmiModClean

e2e-mod-list: e2e-build
	go test -v ./tests/e2e -run AmiModList

e2e-mod-get: e2e-build
	go test -v ./tests/e2e -run AmiModGet

e2e-mod-sum: e2e-build
	go test -v ./tests/e2e -run AmiModSum

e2e-mod-update: e2e-build
	go test -v ./tests/e2e -run AmiModUpdate

# List packages and files under src/ missing test coverage patterns.
# - Reports packages with zero *_test.go files.
# - Reports .go files without a matching *_test.go sibling (same basename).
test-hotspots:
	@echo "Scanning src/ for test coverage hotspots..." >&2
	@# Packages with no tests
	@find src -type d | while read d; do \
	  c_go=$$(ls "$$d"/*.go 2>/dev/null | wc -l | tr -d ' '); \
	  c_test=$$(ls "$$d"/*_test.go 2>/dev/null | wc -l | tr -d ' '); \
	  if [ "$$c_go" != "0" ] && [ "$$c_test" = "0" ]; then \
	    echo "NO_TESTS  $$d"; \
	  fi; \
	 done
	@# Files with no paired *_test.go
	@find src -type f -name "*.go" ! -name "*_test.go" | while read f; do \
	  base=$$(basename "$$f" .go); \
	  dir=$$(dirname "$$f"); \
	  if [ ! -f "$$dir/$${base}_test.go" ]; then \
	    echo "MISSING_PAIR  $$f  (expect: $${dir}/$${base}_test.go)"; \
	  fi; \
	 done | sed 's#//.*$$##'

examples:
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
