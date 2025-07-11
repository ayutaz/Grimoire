# End-to-End Tests

This directory contains end-to-end (E2E) tests for the Grimoire project.

## Test Files

- `e2e_test.go` - Main E2E test setup and utilities
- `e2e_calculator_test.go` - Tests for calculator functionality
- `e2e_loop_test.go` - Tests for loop functionality

## Running Tests

To run all E2E tests:
```bash
go test ./test/...
```

To run specific tests:
```bash
go test ./test -run TestCalculator
```

## Test Structure

Each test:
1. Builds the Grimoire binary
2. Prepares test images
3. Runs the binary with test images
4. Validates the output

## Adding New Tests

1. Create a new test file following the pattern `e2e_<feature>_test.go`
2. Use the test helpers from `e2e_test.go`
3. Ensure proper cleanup of temporary files