name := "containertui"

# Print help message
help:
    #!/bin/bash
    just --list

# Build program binary
build:
    #!/bin/bash
    mkdir -p bin
    go build -ldflags="-s -w" -trimpath -o bin/$({{ name }}) ./cmd

# Install program binary
install:
    #!/bin/bash
    go install ./cmd

# Run the program
run args="":
    #!/bin/bash
    go run ./cmd {{ args }}

# Run program tests
test:
    #!/bin/bash
    go test --cover -parallel=1 -v -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out | sort -rnk3

# Clean development environment
clean:
    @rm -rf coverage.out bin/

# Display test coverage
cover:
    go test -v -race $(shell go list ./... | grep -v /vendor/) -v -coverprofile=coverage.out
    go tool cover -func=coverage.out

# Format program files
fmt:
    gofmt -w -s -l .

# Lint program files
lint:
    golangci-lint run

# Create a test container which logs the date every second.
create-test-container quantity="1":
    #!/bin/bash
    for i in $(seq 1 {{ quantity }}); do
      docker run -d alpine sh -c "while true; do date; sleep 1; done"
    done
