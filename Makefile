
all: ci

ci: lint test build

lint:
	golangci-lint run

test:
	go test -count=20 ./...

build:
	go build -o build/cpngo cmd/main.go

.PHONY: build
