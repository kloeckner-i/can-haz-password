SRC = $(shell find . -type f -name '*.go')

run: $(SRC)
	@go run cmd/main.go

test: $(SRC)
	@go test ./...

lint: $(SRC)
	@gofmt -s -l -w $^
	@goimports -w $^
	@golint ./...
	@golangci-lint run --enable-all -D gomnd -D funlen ./...

clean:
	@go clean -testcache

.PHONY: run test lint clean