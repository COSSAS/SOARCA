.PHONY: all test integration-test ci-test clean build docker run pre-docker-build swagger sbom

BINARY_NAME=soarca
DIRECTORY = $(sort $(dir $(wildcard ./test/*/)))
VERSION = $(shell git describe --tags --dirty)
BUILDTIME := $(shell  date '+%Y-%m-%dT%T%z')

GOLDFLAGS += -X main.Version=$(VERSION)
GOLDFLAGS += -X main.Buildtime=$(BUILDTIME)
GOFLAGS = -ldflags "$(GOLDFLAGS)"

# This creates the swagger.json and swagger.yaml files in api/. \
# These can be copied to docs/static/openapi to provide convenient api documentation
# The same overview of the build can be viewed at /swagger/index.html, including for local build.
swagger:
	mkdir -p api
	swag init -o api -d cmd/soarca/,pkg/models/api,pkg/models/cacao,pkg/models/manual,pkg/api -g main.go

lint: swagger
	
	golangci-lint run --max-same-issues 0 --timeout 5m -v  

build: swagger
	CGO_ENABLED=0 go build -o ./build/soarca $(GOFLAGS) ./cmd/soarca/main.go

test: swagger
	go test ./pkg/... -v
	go test ./internal/... -v

integration-test: swagger
	go test ./test/integration/... -v

ci-test: test integration-test

clean:
	rm -rf build/soarca* build/main
	rm -rf bin/*

compile: swagger
	echo "Compiling for every OS and Platform"
	
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/${BINARY_NAME}-${VERSION}-linux-amd64 $(GOFLAGS) cmd/soarca/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/${BINARY_NAME}-${VERSION}-darwin-arm64 $(GOFLAGS) cmd/soarca/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/${BINARY_NAME}-${VERSION}-windows-amd64 $(GOFLAGS) cmd/soarca/main.go

sbom: swagger
	echo "Generating SBOMs"
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 cyclonedx-gomod app -main cmd/soarca -json -licenses -output bin/${BINARY_NAME}-${VERSION}-linux-amd64.bom.json
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 cyclonedx-gomod app -main cmd/soarca -json -licenses -output bin/${BINARY_NAME}-${VERSION}-darwin-amd64.bom.json
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 cyclonedx-gomod app -main cmd/soarca -json -licenses -output bin/${BINARY_NAME}-${VERSION}-windows-amd64.bom.json

pre-docker-build: swagger
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/${BINARY_NAME}-${VERSION}-linux-amd64 $(GOFLAGS) cmd/soarca/main.go

docker: pre-docker-build
	docker build --no-cache -t soarca:${VERSION}  --build-arg="VERSION=${VERSION}" .

run: docker
	GIT_VERSION=${VERSION} docker compose up --build --force-recreate -d

all: build
