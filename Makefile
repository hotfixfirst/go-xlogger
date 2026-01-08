.PHONY: help test test-coverage test-coverage-html test-race lint fmt vet build clean \
        example-basic example-trace example-all

.DEFAULT_GOAL := help

# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run

# Coverage
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## test: Run all tests
test:
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)

## test-coverage-html: Run tests and generate HTML coverage report
test-coverage-html: test-coverage
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

## test-race: Run tests with race detector
test-race:
	$(GOTEST) -v -race ./...

## lint: Run golangci-lint (requires golangci-lint installed)
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: brew install golangci-lint" && exit 1)
	golangci-lint run ./...

## fmt: Format code
fmt:
	$(GOFMT) ./...

## vet: Run go vet
vet:
	$(GOVET) ./...

## build: Build the package
build:
	$(GOBUILD) ./...

## check: Run fmt, vet, and test
check: fmt vet test

## ci: Run all CI checks (fmt, vet, lint, test with race)
ci: fmt vet lint test-race

## clean: Remove generated files
clean:
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	$(GOCMD) clean -cache -testcache

## example-basic: Run basic example
example-basic:
	@echo "=== Running Basic Example ==="
	$(GORUN) ./_examples/basic/main.go

## example-trace: Run trace example
example-trace:
	@echo "=== Running Trace Example ==="
	$(GORUN) ./_examples/trace/main.go

## example-all: Run all examples
example-all: example-basic example-trace
