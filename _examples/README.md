# Examples

Runnable examples demonstrating `go-xlogger` features.

## Table of Contents

| Example | Description | Run |
| ------- | ----------- | --- |
| [basic](./basic/) | Basic logger usage with config | `cd basic && go run main.go` |
| [trace](./trace/) | Trace context for request tracking | `cd trace && go run main.go` |

## Quick Start

```bash
# Clone the repository
git clone https://github.com/hotfixfirst/go-xlogger.git
cd go-xlogger/_examples

# Run a specific example
cd basic && go run main.go
```

## Adding New Examples

When adding a new feature example:

1. Create a new directory: `_examples/{feature}/`
2. Add `main.go` with runnable code
3. Add `README.md` with documentation
4. Update this file's table of contents
