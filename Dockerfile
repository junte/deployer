FROM golang:1.15.0 

ENV CGO_ENABLED=0

WORKDIR /go/src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN VERSION=$(cat VERSION) \
    && GOOS=linux GO111MODULE=on go build -i -a -installsuffix cgo -ldflags="-X 'main.Version=v${VERSION}'"
