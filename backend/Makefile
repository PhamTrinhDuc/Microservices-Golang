.PHONY: run build test lint

run:
	go run cmd/app.go

build:
	go build -o bin/api cmd/app.go

test:
	go test -v ./...

lint:
	golangci-lint run
