.PHONY: build run test clean docker-build docker-run install dev init

# Variables
BINARY_NAME=dockrune
MAIN_PATH=./cmd/dockrune
VERSION=$(shell git describe --tags --always --dirty)
BUILD_FLAGS=-ldflags="-X main.version=${VERSION}"

# Build the binary and dashboard
build: build-dashboard
	go build ${BUILD_FLAGS} -o ${BINARY_NAME} ${MAIN_PATH}

# Build dashboard
build-dashboard:
	cd dashboard && npm install && npm run build

# Run the server
run: build
	./${BINARY_NAME} serve

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f ${BINARY_NAME}
	rm -f coverage.out coverage.html
	rm -rf dist/

# Install dependencies
deps:
	go mod download
	go mod tidy

# Development mode with hot reload
dev:
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air

# Initialize configuration
init: build
	./${BINARY_NAME} init

# Docker commands
docker-build:
	docker build -t dockrune:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o dist/dockrune-linux-amd64 ${MAIN_PATH}
	GOOS=linux GOARCH=arm64 go build ${BUILD_FLAGS} -o dist/dockrune-linux-arm64 ${MAIN_PATH}
	GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -o dist/dockrune-darwin-amd64 ${MAIN_PATH}
	GOOS=darwin GOARCH=arm64 go build ${BUILD_FLAGS} -o dist/dockrune-darwin-arm64 ${MAIN_PATH}

# Install locally
install: build
	sudo mv ${BINARY_NAME} /usr/local/bin/

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Lint code
lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

# Generate Go documentation
docs:
	@which godoc > /dev/null || go install golang.org/x/tools/cmd/godoc@latest
	@echo "Starting godoc server on http://localhost:6060"
	godoc -http=:6060

# Quick start for development
quickstart: deps init run

# Help
help:
	@echo "dockrune - Self-hosted deployment daemon"
	@echo ""
	@echo "Usage:"
	@echo "  make build        - Build the binary"
	@echo "  make run          - Build and run the server"
	@echo "  make test         - Run tests"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run with docker-compose"
	@echo "  make init         - Initialize configuration"
	@echo "  make dev          - Run in development mode with hot reload"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make help         - Show this help"