# Variables
APP_NAME := talis
GO_FILES := $(shell find . -name "*.go" -type f)
NIX_FILES := $(shell find . -name "*.nix" -type f)
PROJECTNAME=$(shell basename "$(PWD)")

# Go commands
GO := go
GOTEST := $(GO) test
GOVET := $(GO) vet
GOFMT := gofmt
GOMOD := $(GO) mod
GOBUILD := $(GO) build

# Build flags
LDFLAGS := -ldflags="-s -w"

## help: Get more info on make commands.
help: Makefile
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
.PHONY: help

## all: Run check-env, lint, test, and build
all: check-env lint test build
.PHONY: all

## build: Build the application
build: 
	@echo "Building $(APP_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/$(APP_NAME)
.PHONY: build

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/
.PHONY: clean

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...
.PHONY: test

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w $(GO_FILES)
.PHONY: fmt

## lint: Run all linters
lint: vet fmt
	@echo "Running linters..."
	golangci-lint run
.PHONY: lint

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...
.PHONY: vet

## tidy: Tidy and verify dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify
.PHONY: tidy

## run: Run the application
run:
	@echo "Running $(APP_NAME)..."
	@go run ./cmd/main.go
.PHONY: run

## check-env: Check required environment variables
check-env:
	@echo "Checking environment variables..."
	@test -n "$(DIGITALOCEAN_TOKEN)" || (echo "Error: DIGITALOCEAN_TOKEN is not set" && exit 1)
	@test -n "$(PULUMI_ACCESS_TOKEN)" || (echo "Error: PULUMI_ACCESS_TOKEN is not set" && exit 1)
	@test -n "$(PULUMI_ORG)" || (echo "Error: PULUMI_ORG is not set" && exit 1)
	@test -f ~/.ssh/id_rsa || (echo "Error: SSH key not found at ~/.ssh/id_rsa" && exit 1)
	@test -f ~/.ssh/id_rsa.pub || (echo "Error: SSH public key not found at ~/.ssh/id_rsa.pub" && exit 1)
.PHONY: check-env

## nix-check: Validate Nix configurations
nix-check:
	@echo "Validating Nix configurations..."
	@for file in $(NIX_FILES); do \
		echo "Checking $$file..."; \
		nix-instantiate --parse "$$file" >/dev/null || exit 1; \
	done
.PHONY: nix-check

## dev-setup: Setup development environment
dev-setup: check-env
	@echo "Setting up development environment..."
	$(GOMOD) download
	$(GOMOD) verify
	@if ! command -v golangci-lint >/dev/null; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
	fi
.PHONY: dev-setup

## build-cli: Build the Talis CLI tool
build-cli:
	@echo "Building Talis CLI..."
	$(GOBUILD) $(LDFLAGS) -o bin/talis ./cmd/cli
.PHONY: build-cli

## db-connect: Connect to the database
db-connect:
	psql postgres://talis:talis@localhost:5432/talis 
.PHONY: db-connect

# # # # # # # # # # # #
# Docker
# # # # # # # # # # # #

## docker-up: Run the application
docker-up:
	@echo "Launching Docker Compose..."
	@docker compose up -d
.PHONY: docker-up

## docker-down: Stop the application
docker-down:
	@echo "Stopping Docker Compose..."
	@docker compose down
.PHONY: docker-down
