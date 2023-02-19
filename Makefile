
all: lint test build

lint: FORCE
	golangci-lint run

test: FORCE
	go test -count=10 ./...

build: FORCE
	go build -o build/cpngo cmd/cli/main.go
	GOOS=js GOARCH=wasm go build -o build/cpngo.wasm cmd/wasm/main.go
	cp build/cpngo.wasm web/
	rm -rf internal/server/web
	cp -r web internal/server/web
	rm -rf internal/server/web/*wasm*
	
serve: FORCE
	MODE=dev go run cmd/cli/main.go serve

.PHONY: FORCE
