.PHONY: all
all: clean dev

dev: fmt test

fmt:
	go fmt ./...

build:
	go build

test:
	go test -tags=test ./...

clean:
	go clean -i -r -testcache -cache

gen-api:
	protoc -I=./api --go_out=. ./api/api.proto