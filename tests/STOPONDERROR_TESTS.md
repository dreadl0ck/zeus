# stopOnError Testing Documentation

This document describes the test commands created to validate the `stopOnError` functionality in ZEUS.

## Overview

The `stopOnError` field allows control over error handling on a per-command basis:
- When set to `true`: Command execution stops immediately on first error
- When set to `false`: Command execution continues even after errors
- When not specified: Uses the global `StopOnError` config setting

## Test Commands

### Basic Tests (Inline Exec)

#### 1. test-stop-on-error
- **Location**: Inline exec in `commands.yml`
- **stopOnError**: `true`
- **Purpose**: Demonstrates that execution stops after a `false` command
- **Expected behavior**: The message "This should not be printed" will NOT appear

#### 2. test-continue-on-error
- **Location**: Inline exec in `commands.yml`
- **stopOnError**: `false`
- **Purpose**: Demonstrates that execution continues after a `false` command
- **Expected behavior**: The message "This should still be printed" WILL appear

### Script-Based Tests

#### 3. test-stop-on-error-chain
- **Location**: `scripts/test-stop-on-error-chain.sh`
- **stopOnError**: `true`
- **Purpose**: Tests stopOnError with multiple operations (file creation, file reading)
- **Expected behavior**:
  - Creates a test file successfully
  - Attempts to read non-existent file (fails)
  - Does NOT execute cleanup code
  - Error messages about cleanup should NOT appear

#### 4. test-continue-on-error-chain
- **Location**: `scripts/test-continue-on-error-chain.sh`
- **stopOnError**: `false`
- **Purpose**: Tests stopOnError=false with multiple operations
- **Expected behavior**:
  - Creates a test file successfully
  - Attempts to read non-existent file (fails but continues)
  - DOES execute cleanup code
  - Test file is cleaned up successfully
  - Success messages appear

### Dependency Tests

#### 5. test-stop-on-error-with-deps
- **Location**: Inline exec in `commands.yml`
- **stopOnError**: `true`
- **Dependencies**: `test-stop-on-error`
- **Purpose**: Tests stopOnError behavior with command dependencies
- **Expected behavior**: Since the dependency will fail, this command should not execute

### Test Suite

#### 6. test-stop-on-error-suite
- **Location**: `scripts/test-stop-on-error-suite.sh`
- **stopOnError**: `false` (to allow all tests to run)
- **Purpose**: Comprehensive automated test suite that validates stopOnError in various scenarios
- **Tests included**:
  1. Basic stopOnError=true behavior
  2. Basic stopOnError=false behavior
  3. Command substitution handling
  4. Pipeline handling
  5. Explicit exit code handling
- **Expected behavior**: Reports pass/fail for each test and provides summary

## Running the Tests

### Run Individual Tests

```bash
# Test stopOnError=true (should stop on error)
zeus test-stop-on-error

# Test stopOnError=false (should continue on error)
zeus test-continue-on-error

# Test with chained operations
zeus test-stop-on-error-chain
zeus test-continue-on-error-chain

# Test with dependencies
zeus test-stop-on-error-with-deps
```

### Run Complete Test Suite

```bash
zeus test-stop-on-error-suite
```

This will run all tests and provide a comprehensive report.

## Script Files

The following script files have been created in `tests/zeus/scripts/`:

1. `test-stop-on-error.sh` - Basic stopOnError=true test
2. `test-continue-on-error.sh` - Basic stopOnError=false test
3. `test-stop-on-error-chain.sh` - Complex stopOnError=true scenario
4. `test-continue-on-error-chain.sh` - Complex stopOnError=false scenario
5. `test-stop-on-error-suite.sh` - Comprehensive automated test suite

All scripts are executable and can be run independently or through ZEUS commands.

## Configuration

The test setup uses the following configuration in `tests/zeus/config.yml`:
- Global `stopOnError: true` (line 27)
- Individual commands can override this setting

## Testing Different Languages

While the current tests focus on bash scripts, the stopOnError functionality works with all supported languages:
- bash/zsh/sh
- python
- ruby
- lua
- javascript
- perl
- go

Each language has its own `FlagStopOnError` that gets applied when the command is created.

## Validation Checklist

When testing stopOnError functionality, verify:

- [ ] Commands with `stopOnError: true` halt on first error
- [ ] Commands with `stopOnError: false` continue after errors
- [ ] Commands without explicit setting use global config
- [ ] Dependencies respect stopOnError settings
- [ ] Error messages are properly displayed
- [ ] Exit codes are handled correctly
- [ ] Multiple operations in sequence behave as expected
- [ ] Works across different scripting languages

## Notes

- The test suite runs with `stopOnError: false` to ensure all tests execute
- Temporary test files are created in `/tmp/` and cleaned up
- Tests use explicit success/failure indicators (✓/✗) for clarity
- All test scripts include detailed comments explaining expected behavior

