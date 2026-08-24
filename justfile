# goshazam task automation

default: test

# Run tests
test:
	go test -v -race ./...

# Run tests with coverage
coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -func=coverage.txt

# Run linter / static analysis
lint:
	go vet ./...

# Tidy dependencies
tidy:
	go mod tidy -diff

# Build CLI or library examples
build:
	mkdir -p bin
	go build -o bin/goshazam ./cmd/goshazam
