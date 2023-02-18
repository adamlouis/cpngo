
all: lint test build

lint:
	golangci-lint run

test: FORCE
	go test -count=10 ./...

build: FORCE
	go build -o build/cpngo cmd/main.go

run: FORCE
	go run .

.PHONY: FORCE
