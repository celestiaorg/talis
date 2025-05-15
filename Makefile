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

## all: Run check-env, lint, test, build, and generate swagger docs
all: check-env lint test build swagger
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
	$(GOTEST) $(TEST_FLAGS) -race -run="$(TEST)" -count=$(COUNT) $(PKG)
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
run: check-env
	@echo "Running $(APP_NAME)..."
	@go run ./cmd/main.go
.PHONY: run

## check-env: Check required environment variables
check-env:
	@echo "Checking environment variables..."
	@# Check if DIGITALOCEAN_TOKEN is set in environment
	@if [ -n "$(DIGITALOCEAN_TOKEN)" ]; then \
		: ; \
	elif grep -q "DIGITALOCEAN_TOKEN=" .env 2>/dev/null; then \
		: ; \
	else \
		echo "Error: DIGITALOCEAN_TOKEN is not set in environment or .env file" && exit 1; \
	fi
	@# Check if SSH keys are present
	@if [ -f ~/.ssh/id_rsa ] && [ -f ~/.ssh/id_rsa.pub ]; then \
		: ; \
	elif [ -f ~/.ssh/id_ed25519 ] && [ -f ~/.ssh/id_ed25519.pub ]; then \
		: ; \
	else \
		echo "Error: No SSH keys found. Please generate either RSA or ED25519 keys" && exit 1; \
	fi
	@# Check if ansible-playbook is installed
	@command -v ansible-playbook >/dev/null || (echo "Error: ansible-playbook executable not found in PATH" && exit 1)
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
	@if ! command -v swag >/dev/null; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
.PHONY: dev-setup

## swagger: Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@if command -v swag >/dev/null; then \
		echo "Using swag to generate documentation..."; \
		swag init -g cmd/main.go -o docs --parseDependency --parseInternal --parseDepth 2 --generatedTime; \
	else \
		echo "Swag command not found. Using existing files..."; \
		echo "To install swag, run: go install github.com/swaggo/swag/cmd/swag@latest"; \
		echo "Then add it to your PATH or use the full path to run it."; \
	fi
.PHONY: swagger

## build-cli: Build the Talis CLI tool
build-cli:
	@echo "Building Talis CLI..."
	$(GOBUILD) $(LDFLAGS) -o bin/talis-cli ./cmd/cli
.PHONY: build-cli

## run-cli: Run the Talis CLI tool
run-cli:
	@echo "Running Talis CLI..."
	@go run ./cmd/cli/main.go $(ARGS)
.PHONY: run-cli

## db-connect: Connect to the database
db-connect:
	psql postgres://talis:talis@localhost:5432/talis 
.PHONY: db-connect

# # # # # # # # # # # #
# Docker
# # # # # # # # # # # #

## docker-up-local: Run the application
docker-up-local:
	@echo "Launching Docker Compose..."
	@docker compose -f docker-compose-local.yml up -d
.PHONY: docker-up-local

## docker-down-local: Stop the application
docker-down-local:
	@echo "Stopping Docker Compose..."
	@docker compose -f docker-compose-local.yml down
.PHONY: docker-down-local

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


# # # # # # # # # # # #
# Kong
# # # # # # # # # # # #

.PHONY: kong-setup
kong-setup:
	# 1. Create the service
	curl -i -X POST http://localhost:8001/services \
	  --data "name=api" \
	  --data "url=http://api:8080"

	# 2. Create the route (for example, for /talis)
	curl -i -X POST http://localhost:8001/services/api/routes \
	  --data "paths[]=/talis" \
	  --data "strip_path=true"

	# 3. Add the key-auth plugin to the service
	curl -i -X POST http://localhost:8001/services/api/plugins \
	  --data "name=key-auth"

	# 4. Create a consumer
	curl -i -X POST http://localhost:8001/consumers \
	  --data "username=talisuser"

	# 5. Create an API key for the consumer
	curl -i -X POST http://localhost:8001/consumers/talisuser/key-auth
