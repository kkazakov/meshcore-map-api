# Agent Development Guide

This guide provides essential information for AI coding agents working in the meshcore-map-api repository.

## Project Overview

**Language**: Go 1.25.5  
**Type**: Lightweight HTTP API server  
**Framework**: Standard library (`net/http`)  
**Architecture**: Single-file monolithic server with validation layer

## Build & Run Commands

### Build
```bash
# Build executable
go build -o server

# Build and run
go run main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a single test
go test -run TestName -v

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting & Formatting
```bash
# Format code (always run before committing)
go fmt ./...

# Vet code for common issues
go vet ./...

# Install and run golangci-lint (recommended)
golangci-lint run
golangci-lint run --fix
```

### Running the Server
```bash
# Development
go run main.go

# Production (using built binary)
./server

# Server runs on port 8080
```

## Code Style Guidelines

### Imports
- Group imports: standard library, then third-party, then internal packages
- Use `goimports` or let your editor auto-organize imports
- Remove unused imports (enforced by compiler)

```go
import (
    "encoding/json"
    "fmt"
    "log"
    
    "github.com/some/package"
    
    "meshcore-map-api/internal/validator"
)
```

### Formatting
- Use `go fmt` - this is non-negotiable
- Tab indentation (enforced by `go fmt`)
- Line length: aim for 100-120 characters (soft limit)

### Types and Structs
- Use PascalCase for exported types: `type RadioInfo struct`
- Use camelCase for unexported types: `type internalConfig struct`
- Always use struct tags for JSON marshaling
- Match JSON field names to external API contracts exactly

```go
type DeviceData struct {
    DeviceID   string  `json:"deviceId"`    // Matches external API
    DeviceName string  `json:"deviceName"`
    RSSI       int     `json:"rssi"`
    SNR        float64 `json:"snr"`
}
```

### Naming Conventions
- **Functions**: camelCase for unexported, PascalCase for exported
- **Variables**: short names in small scopes (`i`, `err`), descriptive in larger scopes
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE for exported
- **Receivers**: one or two letters, consistent throughout type (`r *Request`, `rr *ReportRequest`)

### Error Handling
- Always check errors, never ignore them
- Return errors up the call stack; handle at appropriate level
- Use `fmt.Errorf()` for formatted error messages
- Include context in errors: `fmt.Errorf("data[%d].timestamp is required", index)`
- Don't use panic except for truly unrecoverable situations

```go
// Good
if err := validateReport(report); err != nil {
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
    return
}

// Bad - ignoring error
_ = validateReport(report)
```

### HTTP Handlers
- Set `Content-Type` header early
- Write status code before response body
- Always `defer r.Body.Close()`
- Return structured JSON errors, not plain text

```go
func handleReport(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    // Validate method
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
        return
    }
    
    // Decode and validate
    var report ReportRequest
    if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON: " + err.Error()})
        return
    }
    defer r.Body.Close()
    
    // Process request
}
```

### Validation Patterns
- Create separate validation functions for each struct type
- Return descriptive errors with field paths: `"data[0].latitude must be between -90 and 90"`
- Validate all required fields before optional ones
- Use early returns for validation failures

### Logging
- Use `log.Printf()` for structured output
- Include relevant context in logs
- Example: `log.Printf("Received valid report from: %s\n", report.Metadata.Name)`

## Testing Guidelines

### Test File Naming
- Place tests in `_test.go` files: `main_test.go`, `validator_test.go`
- Use package name with `_test` suffix or same package for white-box testing

### Test Function Naming
```go
func TestValidateRadioInfo(t *testing.T) { }
func TestHandleReport_InvalidJSON(t *testing.T) { }
func TestHandleReport_Success(t *testing.T) { }
```

### Writing Tests
```go
func TestValidateDeviceData_ValidInput(t *testing.T) {
    data := DeviceData{
        DeviceID:   "abc123",
        DeviceName: "Test Device",
        Timestamp:  "2026-01-16T21:41:52.615226",
        Latitude:   42.6674757,
        Longitude:  23.2714001,
        ScanSource: "active_ping_response",
    }
    
    err := validateDeviceData(data, 0)
    if err != nil {
        t.Errorf("Expected no error, got: %v", err)
    }
}
```

## Common Patterns

### Validation Chain
```go
if err := validateMetadata(report.Metadata); err != nil {
    return err
}
if err := validateData(report.Data); err != nil {
    return err
}
```

### JSON Response Helpers
```go
type ErrorResponse struct {
    Error string `json:"error"`
}

// Success: {"status": "success"}
// Error: {"error": "descriptive message"}
```

## Development Workflow

1. **Before starting**: Run `go fmt ./...` and `go vet ./...`
2. **During development**: Write tests alongside code
3. **Before committing**: 
   - Run `go fmt ./...`
   - Run `go vet ./...`
   - Run `go test ./...`
   - Build successfully: `go build`
4. **Keep it simple**: Prefer standard library over dependencies

## Project-Specific Notes

- This is a minimal API server using only Go's standard library
- Port 8080 is hardcoded (environment variables can be added if needed)
- All validation returns descriptive errors with field paths
- Timestamp validation accepts multiple ISO8601/RFC3339 formats
- No database layer currently - handlers just log and acknowledge

## General instructions

- Don't create any comments. Keep the code clean.
- Don't add emojis anywhere, including documentation.