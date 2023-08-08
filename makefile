build:
	go build -o build/soarca main.go

test:
	go test internal/lib1/lib1_test.go

clean:
	rm -rf build/soarca* build/main
	rm -rf bin/*

compile:
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=amd64 go build -o bin/main-linux-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/main-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -o bin/main-windows-amd64 main.go

all: build