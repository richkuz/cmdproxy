# Local Go toolchain (optional)
GO ?= go

# Docker image for reproducible builds when `go` is not installed on the host
DOCKER_GO_IMAGE ?= golang:1.22-bookworm
DOCKER_RUN = docker run --rm -v "$(CURDIR):/src" -w /src $(DOCKER_GO_IMAGE)

.PHONY: all build test vet fmt docker-test docker-build dist-darwin clean

all: test build

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test: vet
	$(GO) test ./... -count=1

build:
	$(GO) build -o cmdproxy .

# Use when Go is not on PATH (same checks as CI)
docker-test:
	$(DOCKER_RUN) sh -c 'gofmt -w . && go vet ./... && go test ./... -count=1'

docker-build:
	$(DOCKER_RUN) sh -c 'gofmt -w . && go build -o cmdproxy .'

# Cross-compile macOS binaries into dist/ (for release; run on any host with Docker)
dist-darwin:
	@mkdir -p dist/darwin-arm64 dist/darwin-amd64
	$(DOCKER_RUN) sh -c 'set -e; gofmt -w . && go vet ./... && go test ./... -count=1 && \
		CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/darwin-arm64/cmdproxy . && \
		CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/darwin-amd64/cmdproxy .'

clean:
	rm -f cmdproxy
