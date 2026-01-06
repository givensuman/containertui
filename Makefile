projectname?=containertui
# Adapted from:
# https://github.com/FalcoSuessgott/golang-cli-template

default: help

.PHONY: help
help: ## list makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## build golang binary
	@mkdir -p bin
	go build -o bin/$(projectname) ./cmd

.PHONY: install
install: ## install golang binary
	go install ./cmd

.PHONY: run
run: ## run the app
	go run ./cmd

.PHONY: test
test: clean ## run tests with coverage
	go test --cover -parallel=1 -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | sort -rnk3

.PHONY: clean
clean: ## clean up environment
	@rm -rf coverage.out bin/

.PHONY: cover
cover: ## display test coverage
	go test -v -race $(shell go list ./... | grep -v /vendor/) -v -coverprofile=coverage.out
	go tool cover -func=coverage.out

.PHONY: fmt
fmt: ## format go files
	gofumpt -w .

.PHONY: lint
lint: ## lint go files
	golangci-lint run
