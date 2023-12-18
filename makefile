.PHONY: all test clean build

BINARY_NAME=soarca
DIRECTORY = $(sort $(dir $(wildcard ./test/*/)))
VERSION = $(shell git describe --tags --dirty)

lint:
	golangci-lint run -v

build:
	swag init
	go build -o ./build/soarca main.go

test:
	go test test/cacao/*_test.go -v
	go test test/logger/*_test.go -v
	go test test/decomposer/*_test.go -v
	go test test/executor/*_test.go -v
	go test test/routes/workflow_api/*_test.go -v
	go test test/routes/trigger_api/*_test.go -v
	go test test/capability/ssh/*_test.go -v
	go test test/capability/http/*_test.go -v


clean:
	rm -rf build/soarca* build/main
	rm -rf bin/*

compile:
	echo "Compiling for every OS and Platform"
	
	swag init
	GOOS=linux GOARCH=amd64 go build -o bin/${BINARY_NAME}-${VERSION}-linux-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/${BINARY_NAME}-${VERSION}-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -o bin/${BINARY_NAME}-${VERSION}-windows-amd64 main.go

all: build

