.PHONY: lint
lint:
	@golangci-lint run

.PHONY: build
build:
	@go build -v -o kv ./cmd/kv

.PHONY: test
test:
	@go test -v ./...