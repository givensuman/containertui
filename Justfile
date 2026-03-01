name := "containertui"

# Print help message
help:
    #!/bin/bash
    just --list

# Note: Demo generation requires a Distrobox environment with vhs, Go, and Docker.
# See distrobox.ini and create-distrobox.sh for setup instructions.
# To enter the distrobox environment, run: distrobox enter containertui-demos

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
    DEBUG=true
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

# Set up Docker test environment for demos
demo-setup:
    #!/bin/bash
    ./demos/setup.sh

# Clean up Docker test environment
demo-cleanup:
    #!/bin/bash
    ./demos/cleanup.sh

# Generate a single demo GIF
demo-single tape: demo-setup
    #!/bin/bash
    vhs demos/{{ tape }}.tape -o assets/demo-{{ tape }}.gif
    @just demo-cleanup

# Generate all demo GIFs
demo: demo-setup
    #!/bin/bash
    echo "Setting up Docker test environment..."
    ./demos/setup.sh
    echo ""
    echo "Generating demo GIFs..."
    for tape in overview containers images volumes networks services; do
        echo "  Generating $tape.gif..."
        vhs demos/$tape.tape -o assets/demo-$tape.gif || echo "Warning: $tape.tape failed"
    done
    echo ""
    echo "Cleaning up Docker test environment..."
    ./demos/cleanup.sh
    echo ""
    echo "Done! Generated GIFs are in assets/"
    ls -lh assets/demo-*.gif

    @just demo-cleanup
