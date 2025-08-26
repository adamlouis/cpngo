
all: lint test build

lint: FORCE
	golangci-lint run

test: FORCE
	go test -count=20 ./...

build: build-go build-web

build-go: FORCE
	go build -o build/cpngo cmd/main.go

serve: FORCE
	MODE=dev go run cmd/cli/main.go serve

.PHONY: FORCE
