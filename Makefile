# Variables
APP_NAME := talis
GO_FILES := $(shell find . -name "*.go" -type f)
PROJECTNAME=$(shell basename "$(PWD)")

# Go commands
GO := go
GOTEST := $(GO) test
GOVET := $(GO) vet
GOFMT := gofmt
GOMOD := $(GO) mod
GOBUILD := $(GO) build

# flags
PKG ?= ./...
TEST ?= .
TEST_FLAGS ?= -v
COUNT ?= 1

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
	$(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/main.go
.PHONY: build

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/
.PHONY: clean

## test: Run tests
# You can specify a package with 'make test PKG=./path/to/package'
# You can specify a test pattern with 'make test TEST=TestName'
# You can specify test flags with 'make test TEST_FLAGS="-v -cover"'
# Examples:
#   make test                         # Run all tests
#   make test PKG=./internal/auth     # Run tests in the auth package
#   make test TEST=TestLogin  # Run tests matching TestLogin
#   make test PKG=./internal/auth TEST=TestLogin  # Run TestLogin in auth package
test:
	@echo "Running tests..."
	$(GOTEST) $(TEST_FLAGS) -run="$(TEST)" -count=$(COUNT) $(PKG)
.PHONY: test

## fmt: Format code
fmt:
	@echo "Formatting go fmt..."
	$(GOFMT) -w $(GO_FILES)
	@echo "--> Formatting golangci-lint"
	@golangci-lint run --fix
.PHONY: fmt

## lint: Run all linters
lint: tidy fmt vet
	@echo "Running linters..."
	@echo "--> Running golangci-lint"
	@golangci-lint run
	@echo "--> Running actionlint"
	@actionlint
	@echo "--> Running yamllint"
	@yamllint --no-warnings .
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

## install-hooks: Install git hooks
install-hooks:
	@echo "Installing git hooks..."
	@git config core.hooksPath .githooks
.PHONY: install-hooks

## dev-setup: Setup development environment
dev-setup: check-env install-hooks
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
	$(GOBUILD) $(LDFLAGS) -o bin/talis-cli ./cmd/cli
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
