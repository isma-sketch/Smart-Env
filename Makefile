BINARY := smart-env
MODULE := github.com/smartenv/smart-env
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build clean test lint cross

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/smart-env

test:
	go test ./...

lint:
	gofmt -l .
	go vet ./...

clean:
	rm -f $(BINARY) smart-env.exe

# Cross-compilation
cross: cross-linux cross-darwin cross-windows

cross-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64 ./cmd/smart-env
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64 ./cmd/smart-env

cross-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-amd64 ./cmd/smart-env
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-arm64 ./cmd/smart-env

cross-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-windows-amd64.exe ./cmd/smart-env
