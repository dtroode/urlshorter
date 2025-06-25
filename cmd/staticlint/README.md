# Static Code Analyzer

A comprehensive static code analyzer for Go projects that combines standard Go analyzers, staticcheck.io analyzers, and custom analyzers.

## Features

- **Standard Go Analyzers**: All analyzers from `golang.org/x/tools/go/analysis/passes`
- **Staticcheck Analyzers**: Security and style analyzers from staticcheck.io
- **Custom Analyzers**: Project-specific analyzers (e.g., osexit for detecting os.Exit() usage)
- **Portable Binary**: Can be used with any Go project without installation

## Installation

### Build from source
```bash
# Clone the repository
git clone <repository-url>
cd urlshorter

# Build the portable binary
go build -o mycheck cmd/staticlint/main.go
```

### Install globally (optional)
```bash
go install ./cmd/staticlint
```

## Usage

### Basic usage
```bash
# Analyze specific files
./mycheck cmd/shortener/main.go

# Analyze entire project
./mycheck ./...

# Analyze specific directories
./mycheck cmd/... internal/...
```

### Portable usage
The binary can be used with any Go project:

```bash
# Analyze another project
./mycheck /path/to/other/project/main.go
./mycheck /path/to/other/project/...

# Use globally installed version
mycheck ./...
```

## Analyzers

### Standard Go Analyzers
- **appends**: Checks correct usage of append()
- **assign**: Checks correctness of assignments
- **atomic**: Checks correct usage of atomic operations
- **bools**: Checks simplification of boolean expressions
- **buildtag**: Checks correctness of build tags
- **cgocall**: Checks CGO function calls
- **composite**: Checks correctness of composite literals
- **copylock**: Checks copying of structs with locks
- **defers**: Checks correct usage of defer
- **errorsas**: Checks correct usage of errors.As
- **fieldalignment**: Checks field alignment in structs
- **httpresponse**: Checks HTTP response handling
- **loopclosure**: Checks closures in loops
- **lostcancel**: Checks lost context cancellations
- **nilfunc**: Checks nil function calls
- **nilness**: Checks nil values
- **printf**: Checks correctness of printf functions
- **shadow**: Checks variable shadowing
- **shift**: Checks correctness of bit shifts
- **structtag**: Checks struct tag correctness
- **tests**: Checks test correctness
- **unreachable**: Checks unreachable code
- **unsafeptr**: Checks unsafe.Pointer usage
- **unusedresult**: Checks unused function results
- **waitgroup**: Checks correct usage of WaitGroup

### Staticcheck Analyzers
- **SA1000-SA1032**: Security analysis checks
- **SA2000-SA2003**: Concurrency checks
- **SA3000-SA3001**: Testing checks
- **SA4000-SA4032**: Performance checks
- **SA5000-SA5012**: Correctness checks
- **SA6000-SA6006**: Style checks
- **SA9001-SA9009**: Miscellaneous checks

### Custom Analyzers
- **osexit**: Detects os.Exit() usage in main() function of main package

## Configuration

The analyzer uses a predefined set of analyzers. To modify the configuration:

1. Edit the `checks` slice in `main.go` to add/remove standard analyzers
2. Modify the `staticchecks` map to change staticcheck analyzer selection
3. Add custom analyzers to the `checks` slice

## Exit Codes

- **0**: No issues found
- **1**: Analysis errors (e.g., compilation errors)
- **3**: Issues found (standard exit code for static analyzers)

## Examples

### Detecting os.Exit() usage
```go
package main

import "os"

func main() {
    os.Exit(1) // This will trigger the osexit analyzer
}
```

Output:
```
main.go:5:2: os call detected
```

### Running on a large project
```bash
./mycheck ./... 2>&1 | grep -v "analysis skipped"
```

## Integration

### CI/CD
```yaml
# GitHub Actions example
- name: Run static analysis
  run: |
    go build -o mycheck cmd/staticlint/main.go
    ./mycheck ./...
```

### Pre-commit hooks
```bash
#!/bin/sh
# .git/hooks/pre-commit
go build -o mycheck cmd/staticlint/main.go
./mycheck ./...
```

## Troubleshooting

### Common issues

1. **"no required module provides package"**: Ensure the target project has a valid `go.mod` file
2. **"analysis skipped due to errors"**: Fix compilation errors in the target project first
3. **Permission denied**: Make sure the binary is executable (`chmod +x mycheck`)

### Debug mode
```bash
# Run with verbose output
./mycheck -debug=fpstv ./...
```

## Contributing

To add new analyzers:

1. Create a new analyzer in the `osexit` package style
2. Add it to the `checks` slice in `main.go`
3. Update this README with documentation

## License

[Add your license information here] 