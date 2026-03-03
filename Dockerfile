FROM golang:1.26

ENV CGO_ENABLED=0

WORKDIR /go/src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev

RUN GOOS=linux GO111MODULE=on CGO_ENABLED=0 \
  go build -a -installsuffix cgo \
  -ldflags="-X 'deployer/src/config.Version=${VERSION}'" \
  -o bin/deployer \
  deployer/src \
  && cp bin/deployer /usr/local/bin/deployer
