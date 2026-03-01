# Dev

## http

```shell
curl -X POST http://localhost:7778 -d "component=app&key=242134321432143214213&text=bla"

# Async deploy
curl -X POST http://localhost:7778 -d "component=app&key=242134321432143214213&text=bla&async=true"

# Async deploy with extra params
curl -X POST http://localhost:7778 -d "component=app&key=242134321432143214213&text=bla&async&param1=1"
```

## client

```shell
go run src/main.go client --url http://localhost:7778 --component app --key 242134321432143214213 --arg text=bla
```
