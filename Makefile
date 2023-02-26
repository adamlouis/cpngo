
all: lint test build

lint: FORCE
	golangci-lint run

test: FORCE
	go test -count=20 ./...

build: build-go build-web

build-go: FORCE
	go build -o build/cpngo cmd/cli/main.go
	GOOS=js GOARCH=wasm go build -o build/cpngo.wasm cmd/wasm/main.go

build-web:
	# update cnpgo.wasm
	cp build/cpngo.wasm docs/
	# update embedded go webpage
	rm -rf internal/server/web
	cp -r docs internal/server/web
	# remove wasm file from embedded go webpage
	rm -rf internal/server/web/*wasm*

serve: FORCE
	MODE=dev go run cmd/cli/main.go serve

.PHONY: FORCE
