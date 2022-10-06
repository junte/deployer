VERSION=$(shell cat VERSION)

build:
	@go build -o bin/deployer deployer/cmd/server

build_test_binary_linux:
	@GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-X 'deployer/internal/config.Version=v${VERSION}'" -o bin/deployer deployer/cmd/server

test:
	@go test -v ./...

tag:
	@git tag -a v${VERSION} -m "v${VERSION}"

lint:
	@golangci-lint run

update_libs:
	@go get -u ./... && go mod tidy

