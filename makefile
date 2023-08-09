.PHONY: all test clean build

BINARY_NAME=soarca
DIRECTORY = $(sort $(dir $(wildcard ./test/*/)))

build:
	go build -o ./build/soarca main.go

test:
	@echo $(sort $(dir $(wildcard ./test/*/)))
	go test internal/cacao/*_test.go

clean:
	rm -rf build/soarca* build/main
	rm -rf bin/*

compile:
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=amd64 go build -o bin/${BINARY_NAME}-linux-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/${BINARY_NAME}-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -o bin/${BINARY_NAME}-windows-amd64 main.go

all: build

