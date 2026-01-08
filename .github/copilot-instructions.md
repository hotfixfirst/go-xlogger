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
├── go.mod
├── go.sum
├── LICENSE
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

## Code Style

### Package Declaration
```go
// Package {packagename} provides {brief description of what the package does}.
package {packagename}
```

**Example:**
```go
// Package xlogger provides utilities for logging.
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

## Placeholders Reference

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `{project}` | Project/repo name | `go-xlogger` |
| `{package}` | Go package name | `xlogger` |
| `{packagename}` | Package name in code | `xlogger` |
| `{org}` | GitHub organization | `hotfixfirst` |
| `{feature}` | Feature name (lowercase) | `duration` |
| `{Feature}` | Feature name (Title Case) | `Duration` |
| `{FunctionName}` | Function name | `ParseDurationToSeconds` |
