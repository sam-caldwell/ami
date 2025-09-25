.PHONY := all clean lint test build

all: build

clean:
	rm -rf ./build
	mkdir -p ./build

lint:
	go vet -v ./...

test:
	go test -v ./...

build: clean
	go build -o build/ami src/main.go

