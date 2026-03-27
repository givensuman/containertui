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

# Run interactive Docker demo with Docker-in-Docker
# Launches a containerized demo environment with sample containers, images, volumes, and networks
demo-interactive:
    #!/bin/bash
    set -e
    
    IMAGE_NAME="containertui-demo:latest"
    CONTAINER_NAME="containertui-demo-$$"
    
    echo "Building demo image..."
    docker build -f Dockerfile.demo -t "$IMAGE_NAME" . --quiet
    
    echo ""
    echo "Launching interactive demo environment..."
    echo "Press Ctrl+C to exit when you're done."
    echo ""
    
    # Run the demo container with DinD capabilities
    docker run --rm -it \
        --privileged \
        --name "$CONTAINER_NAME" \
        --tmpfs /run \
        --tmpfs /tmp \
        "$IMAGE_NAME" || true
    
    echo ""
    echo "Demo environment has been cleaned up."

# Run interactive Docker-in-Docker demo
# This is the primary demo command - launches a containerized environment
demo: demo-interactive

# Generate all demo GIFs (same environment as the demo container)
demo-gifs:
    #!/bin/bash
    set -e
    
    echo "Building containertui binary..."
    go build -o /tmp/containertui ./cmd
    
    echo "Ensuring demo Docker image is built..."
    docker build -f Dockerfile.demo -t containertui-demo:latest . --quiet
    
    echo "Setting up demo Docker environment..."
    ./demos/setup.sh
    
    echo ""
    echo "Generating demo GIFs using VHS..."
    cd demos
    
    for tape in overview containers images volumes networks services; do
        echo "  Generating $tape.gif..."
        vhs "$tape.tape" -o "../assets/demo-$tape.gif" || echo "Warning: $tape.tape failed"
    done
    
    cd ..
    echo ""
    echo "Cleaning up demo Docker environment..."
    ./demos/cleanup.sh
    
    echo ""
    echo "Done! Generated GIFs are in assets/"
    ls -lh assets/demo-*.gif
