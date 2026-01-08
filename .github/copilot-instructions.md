# Copilot Instructions for Go SDK Projects

## Project Structure

When creating or modifying this Go SDK, follow this structure:

```
{project}/
├── .github/
│   └── copilot-instructions.md   # This file
├── _examples/
│   ├── README.md                 # Table of contents for all examples
│   └── {feature}/
│       ├── main.go               # Runnable example
│       └── README.md             # Example documentation
├── {feature}.go                  # Implementation
├── {feature}_test.go             # Unit tests (required)
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE
├── Makefile                      # Build and development commands
└── README.md                     # Main documentation
```

## Rules

### 1. File Naming
- Implementation: `{feature}.go` (e.g., `duration.go`, `timezone.go`)
- Tests: `{feature}_test.go` (e.g., `duration_test.go`)
- Examples: `_examples/{feature}/main.go`

### 2. When Creating New Feature
Always create these files together:
- [ ] `{feature}.go` - Main implementation
- [ ] `{feature}_test.go` - Unit tests (minimum 80% coverage)
- [ ] `_examples/{feature}/main.go` - Working example
- [ ] `_examples/{feature}/README.md` - How to run example

### 3. When Updating Feature
- [ ] Update `{feature}.go`
- [ ] Add/update tests in `{feature}_test.go`
- [ ] Update example in `_examples/{feature}/main.go`
- [ ] Update documentation

### 4. Documentation Updates
When adding new feature, update:
- [ ] Root `README.md` - Add to packages table and create section
- [ ] `_examples/README.md` - Add to table of contents

## Markdown Style

Follow these rules to avoid markdown linting warnings:

### Table Separators
Use spaces around pipes in separator rows:

```markdown
<!-- ✅ Good -->
| Column 1 | Column 2 |
| -------- | -------- |
| value    | value    |

<!-- ❌ Bad -->
| Column 1 | Column 2 |
|----------|----------|
| value    | value    |
```

### Code Block Language
Always specify language for fenced code blocks:

```markdown
<!-- ✅ Good -->
```text
Sample output here
```

<!-- ❌ Bad -->
```
Sample output here
```
```

### Headings
- Use unique heading names (no duplicates)
- Add blank line before and after headings

### Lists
- Add blank line before and after lists

```markdown
<!-- ✅ Good -->
### Section

- Item 1
- Item 2

### Next Section

<!-- ❌ Bad -->
### Section
- Item 1
- Item 2
### Next Section
```

## Code Style

### Package Declaration
```go
// Package {packagename} provides {brief description of what the package does}.
package {packagename}
```

**Example:**
```go
// Package xlogger provides advanced logging functionalities.
package xlogger
```

### Function Documentation
Follow Go standard documentation format:

```go
// {FunctionName} {brief description starting with verb}.
//
// {Detailed description if needed, explaining:}
// - What the function does
// - Parameters and their expected values
// - Return values
//
// Supported formats: {list if applicable}
//
// Example:
//
//	result, err := {FunctionName}(input)
//	// result = expected output
func {FunctionName}(param Type) (ReturnType, error) {
```

**Example:**
```go
// ParseValue parses a string value and returns the parsed result.
//
// The input string must be in a valid format. Returns an error
// if the format is invalid or the value cannot be parsed.
//
// Supported formats: "type1", "type2", "type3"
//
// Example:
//
//	result, err := ParseValue("type1")
//	// result = 100
func ParseValue(input string) (int64, error) {
```

### Error Handling
- Return errors, don't panic (except `Must*` functions)
- Use descriptive error messages
- Wrap errors with context when needed

```go
// Good
return 0, fmt.Errorf("invalid format: %s", input)

// With context
return 0, fmt.Errorf("parse config: %w", err)
```

### Testing
```go
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    int64
        wantErr bool
    }{
        {"valid case", "input1", 100, false},
        {"invalid case", "invalid", 0, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Example Templates

### _examples/{feature}/main.go
```go
// Package main demonstrates the usage of the {package} {feature} functionality.
package main

import (
    "fmt"
    "github.com/{org}/{project}"
)

func main() {
    fmt.Println("=== {Feature} Examples ===")
    fmt.Println()
    
    // Example 1: Basic usage
    fmt.Println("1. Basic Usage")
    fmt.Println("--------------")
    // ... examples
    
    fmt.Println()
    fmt.Println("=== End of Examples ===")
}
```

### _examples/{feature}/README.md
````markdown
# {Feature} Example

This example demonstrates the `{package}` {feature} functionality.

## Run

```bash
cd _examples/{feature}
go run main.go
```

## Features Demonstrated

| # | Feature | Function |
|---|---------|----------|
| 1 | ... | `FunctionName()` |

## Sample Output

```
=== {Feature} Examples ===
...
```
````

## README.md Structure

### Root README.md
````markdown
# {project}

{Brief description of what the project does.}

## Installation

```bash
go get github.com/{org}/{project}
```

Or with a specific version:

```bash
go get github.com/{org}/{project}@v1.0.0
```

## Quick Start

```go
import "{org}/{project}"

// Basic example
result, err := {project}.FunctionName("input")
```

## Packages

| Package | Description | Documentation |
|---------|-------------|---------------|
| [{Feature1}](#{feature1}) | {description} | [Examples](./_examples/{feature1}/) |

## {Feature1}

### Functions

| Function | Description |
|----------|-------------|
| `FunctionName()` | {description} |

### Examples

```go
// example code
```

## Examples

See the [_examples](./_examples/) directory for runnable examples.

## Contributing

{Contributing guidelines}

## License

{License information}
````

### _examples/README.md
````markdown
# Examples

Runnable examples demonstrating `{project}` features.

## Table of Contents

| Example | Description | Run |
|---------|-------------|-----|
| [{feature}](./{feature}/) | {description} | `cd {feature} && go run main.go` |

## Quick Start

```bash
# Clone the repository
git clone https://github.com/{org}/{project}.git
cd {project}/_examples

# Run a specific example
cd {feature} && go run main.go
```

## Adding New Examples

When adding a new feature example:

1. Create a new directory: `_examples/{feature}/`
2. Add `main.go` with runnable code
3. Add `README.md` with documentation
4. Update this file's table of contents
````

## Makefile

Every Go SDK project should have a Makefile with standard targets.

### Required Targets

```makefile
.PHONY: help test test-coverage test-race lint fmt vet build clean \
        example-basic example-chaining example-wrapping example-all

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
```

### Example Targets

When adding new examples, add corresponding make targets:

```makefile
## example-{feature}: Run {feature} example
example-{feature}:
	@echo "=== Running {Feature} Example ==="
	$(GORUN) ./_examples/{feature}/main.go

## example-all: Run all examples
example-all: example-{feature1} example-{feature2}
```

### Standard Targets Reference

| Target | Description |
| ------ | ----------- |
| `help` | Show available commands |
| `test` | Run all tests |
| `test-coverage` | Run tests with coverage report |
| `test-coverage-html` | Generate HTML coverage report |
| `test-race` | Run tests with race detector |
| `lint` | Run golangci-lint |
| `fmt` | Format code |
| `vet` | Run go vet |
| `build` | Build the package |
| `check` | Run fmt, vet, and test |
| `ci` | Run all CI checks |
| `clean` | Remove generated files |
| `example-{name}` | Run specific example |
| `example-all` | Run all examples |

## Placeholders Reference

| Placeholder | Description | Example |
| ----------- | ----------- | ------- |
| `{project}` | Project/repo name | `go-xlogger` |
| `{package}` | Go package name | `xlogger` |
| `{packagename}` | Package name in code | `xlogger` |
| `{org}` | GitHub organization | `hotfixfirst` |
| `{feature}` | Feature name (lowercase) | `duration` |
| `{Feature}` | Feature name (Title Case) | `Duration` |
| `{FunctionName}` | Function name | `ParseDurationToSeconds` |

## Code Generation Rules

### Creating New Files vs Editing Existing Files

**Creating new files** - Do NOT use `// ...existing code...` or similar markers:

```go
// ✅ Good - Complete file content
package main

import "fmt"

func main() {
    fmt.Println("Hello")
}
```

```go
// ❌ Bad - Using existing code marker in new file
// ...existing code...
func main() {
    fmt.Println("Hello")
}
```

**Editing existing files** - Use context to show where changes go:

```go
// ✅ Good - Show surrounding context
func existingFunction() {
    // existing code
}

func newFunction() {
    // new code here
}
```

### Summary

| Scenario | Use `// ...existing code...` |
| -------- | ---------------------------- |
| Creating new file | ❌ Never use |
| Editing existing file | ⚠️ Avoid, use actual context instead |
| Showing code examples | ✅ OK for documentation |
