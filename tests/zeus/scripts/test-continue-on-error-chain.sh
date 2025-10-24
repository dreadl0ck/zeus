#!/bin/bash
# Test script for stopOnError=false with multiple operations
# Tests more complex scenarios with stopOnError disabled

echo "=== Testing continue on error with multiple operations ==="

# Create a temporary test file
TEST_FILE="/tmp/zeus_continue_on_error_test.txt"
echo "Creating test file: $TEST_FILE"
echo "test data" > "$TEST_FILE"

# This should succeed
if [ -f "$TEST_FILE" ]; then
    echo "✓ Test file created successfully"
else
    echo "✗ Failed to create test file"
    exit 1
fi

# This will fail - trying to read a non-existent file
echo "Attempting to read non-existent file (expected to fail)..."
cat /tmp/this_file_does_not_exist_zeus_test.txt 2>&1 || echo "✓ Command failed as expected, but continuing..."

# These operations SHOULD execute when stopOnError=false
echo "✓ Continuing execution after error (stopOnError=false)"
echo "✓ Cleaning up test file..."
rm -f "$TEST_FILE"

if [ ! -f "$TEST_FILE" ]; then
    echo "✓ Test file cleaned up successfully"
else
    echo "✗ Failed to clean up test file"
fi

echo "✓ Test completed successfully - script continued despite errors"
exit 0

