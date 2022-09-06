VERSION=$(shell cat VERSION)

build:
	@go build -o bin/deployer deployer/cmd/server

test:
	@go test -v ./...

tag:
	@git tag -a v${VERSION} -m "v${VERSION}"

lint:
	@golangci-lint run
