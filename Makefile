.PHONY: build test clean install

BINARY := bin/term2mcp

build:
	go build -o $(BINARY) .

test:
	go test ./...

clean:
	rm -rf bin/

install:
	go install .

run:
	go run .

fmt:
	go fmt ./...
	goimports -w .
