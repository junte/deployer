VERSION=$(shell cat VERSION)
GIT_VERSION=$(shell git describe --tags --always --dirty)

build:
	go build -ldflags="-X 'deployer/src/config.Version=$(GIT_VERSION)'" -o bin/deployer deployer/src

build-linux:
	GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-X 'deployer/src/config.Version=$(GIT_VERSION)'" -o bin/deployer deployer/src

build-all:
	@for platform in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		echo "building $$os/$$arch..."; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -a -installsuffix cgo \
			-ldflags="-X 'deployer/src/config.Version=$(GIT_VERSION)'" \
			-o bin/deployer-$$os-$$arch deployer/src; \
	done

test:
	go test -v ./...

tag:
	git tag -a v${VERSION} -m "v${VERSION}"

lint:
	golangci-lint run
