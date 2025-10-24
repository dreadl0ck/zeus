#!/bin/bash
# Test script for stopOnError with multiple operations
# Tests more complex scenarios with stopOnError enabled

echo "=== Testing stopOnError with multiple operations ==="

# Create a temporary test file
TEST_FILE="/tmp/zeus_stop_on_error_test.txt"
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
echo "Attempting to read non-existent file..."
cat /tmp/this_file_does_not_exist_zeus_test.txt

# These operations should NOT execute when stopOnError=true
echo "ERROR: Cleanup should not happen - stopOnError should have stopped execution"
rm -f "$TEST_FILE"
echo "ERROR: This message should never appear"

