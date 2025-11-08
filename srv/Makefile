
## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'


## test: run all tests
.PHONY: test
test:
	go test -v ./...

## cover: run all tests and display coverage
.PHONY: cover
cover:
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

## tidy: tidy modfiles and format .go files
.PHONY: tidy
tidy:
	go mod tidy -v
	go fmt ./...

dep:
	go mod download

vet:
	go vet

## build: build the application
.PHONY: build
build:
	go build
