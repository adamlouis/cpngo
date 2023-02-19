
all: lint test build

lint: FORCE
	golangci-lint run

test: FORCE
	go test -count=10 ./...

build: FORCE
	go build -o build/cpngo cmd/main.go


serve: FORCE
	go run cmd/main.go serve

.PHONY: FORCE
