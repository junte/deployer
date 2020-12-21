build:
	@go build -o bin/deployer deployer/cmd/server

test:
	@go test -v ./...