FROM golang:1.15

ENV CGO_ENABLED=0

WORKDIR /go/src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN VERSION=$(cat VERSION) \
    && GOOS=linux GO111MODULE=on CGO_ENABLED=0 \
        go build -a -installsuffix cgo \
        -ldflags="-X 'deployer/internal/core.Version=v${VERSION}'" \
        -o bin/deployer \
        deployer/cmd/server
