# Variables
BINARY_NAME=tiny-coder
BUILD_DIR=bin
VERSION=0.0.1

.PHONY: all help tidy fmt vet test build clean run run-precommit

## all: Run standard quality checks and build the binary
all: tidy fmt vet test build

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'

## tidy: Add missing and remove unused modules
tidy:
	go mod tidy

## fmt: Format all Go source files
fmt:
	go fmt ./...

## vet: Examine Go source code and report suspicious constructs
vet:
	go vet ./...

## test: Run all unit tests
test:
	go test -v -race ./...

## build: Build the Go binary for the current OS/Architecture
build:
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)

## run: Build and execute the application
run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

## run-precommit: Run pre-commit hooks via Lefthook
run-precommit:
	go tool -modfile=tools/tools.mod lefthook run pre-commit