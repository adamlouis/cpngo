
ci: tidy lint test build

lint:
	golangci-lint run

test:
	go test -count=20 ./...

tidy:
	go mod tidy

build:
	go build -o build/cpngo cmd/main.go

.PHONY: build
