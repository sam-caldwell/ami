.PHONY: all clean lint test build examples e2e-build e2e-test

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

examples:
	@if [ ! -x build/ami ]; then \
	  echo "ami binary not found. Build it first: 'go build -o build/ami ./src/cmd/ami'"; \
	  exit 1; \
	fi
	# Build all example workspaces and stage their outputs under build/examples/
	rm -rf build/examples
	mkdir -p build/examples
	set -e; \
	for d in examples/*; do \
	  if [ -f "$$d/ami.workspace" ]; then \
	    echo "Building example: $$d"; \
	    (cd "$$d" && ../../build/ami build --verbose); \
	    name=$$(basename "$$d"); \
	    mkdir -p "build/examples/$$name"; \
	    if [ -d "$$d/build" ]; then \
	      cp -R "$$d/build/." "build/examples/$$name/"; \
	    fi; \
	  fi; \
	done
