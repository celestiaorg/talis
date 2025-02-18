# Variables
APP_NAME := talis
GO_FILES := $(shell find . -name "*.go" -type f)
NIX_FILES := $(shell find . -name "*.nix" -type f)

# Go commands
GO := go
GOTEST := $(GO) test
GOVET := $(GO) vet
GOFMT := gofmt
GOMOD := $(GO) mod
GOBUILD := $(GO) build

# Build flags
LDFLAGS := -ldflags="-s -w"

# TODO: add them to the right place
.PHONY: all build clean test fmt lint vet tidy run help check-env migrate migrate-down migrate-force migrate-version db-connect build-cli

## help: Get more info on make commands.
help: Makefile
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
.PHONY: help

# Default target
all: check-env lint test build

# Build the application
build: 
	@echo "Building $(APP_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/$(APP_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w $(GO_FILES)

# Run all linters
lint: vet fmt
	@echo "Running linters..."
	golangci-lint run

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Tidy and verify dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	go run ./cmd/main.go

# Check required environment variables
check-env:
	@echo "Checking environment variables..."
	@test -n "$(DIGITALOCEAN_TOKEN)" || (echo "Error: DIGITALOCEAN_TOKEN is not set" && exit 1)
	@test -n "$(PULUMI_ACCESS_TOKEN)" || (echo "Error: PULUMI_ACCESS_TOKEN is not set" && exit 1)
	@test -n "$(PULUMI_ORG)" || (echo "Error: PULUMI_ORG is not set" && exit 1)
	@test -f ~/.ssh/id_rsa || (echo "Error: SSH key not found at ~/.ssh/id_rsa" && exit 1)
	@test -f ~/.ssh/id_rsa.pub || (echo "Error: SSH public key not found at ~/.ssh/id_rsa.pub" && exit 1)

# Validate Nix configurations
nix-check:
	@echo "Validating Nix configurations..."
	@for file in $(NIX_FILES); do \
		echo "Checking $$file..."; \
		nix-instantiate --parse "$$file" >/dev/null || exit 1; \
	done

# Development setup
dev-setup: check-env
	@echo "Setting up development environment..."
	$(GOMOD) download
	$(GOMOD) verify
	@if ! command -v golangci-lint >/dev/null; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
	fi

# Build the Talis CLI tool
build-cli:
	@echo "Building Talis CLI..."
	$(GOBUILD) $(LDFLAGS) -o bin/talis ./cmd/cli

# Default target
.DEFAULT_GOAL := help 

migrate:
	go run cmd/migrate/main.go

migrate-down:
	go run cmd/migrate/main.go -down

migrate-force:
	go run cmd/migrate/main.go -force $(version)

migrate-version:
	go run cmd/migrate/main.go -steps 0

db-connect:
	psql postgres://talis:talis@localhost:5432/talis 
