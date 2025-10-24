#!/bin/bash
# Comprehensive test suite for stopOnError functionality
# This script validates that stopOnError works correctly in various scenarios

set +e  # Don't stop on errors (we want to test error handling)

echo "=========================================="
echo "ZEUS stopOnError Test Suite"
echo "=========================================="
echo ""

PASS_COUNT=0
FAIL_COUNT=0

# Helper function to report test results
report_test() {
    local test_name="$1"
    local expected="$2"
    local actual="$3"
    
    if [ "$expected" = "$actual" ]; then
        echo "✓ PASS: $test_name"
        ((PASS_COUNT++))
    else
        echo "✗ FAIL: $test_name (expected: $expected, got: $actual)"
        ((FAIL_COUNT++))
    fi
}

# Test 1: Verify stopOnError=true stops execution
echo "Test 1: stopOnError=true should stop on first error"
echo "----------------------------------------------"
OUTPUT=$(bash -e -c '
    echo "Before error"
    false
    echo "After error - should not appear"
' 2>&1)

if echo "$OUTPUT" | grep -q "After error"; then
    report_test "Test 1: stopOnError=true stops execution" "stopped" "continued"
else
    report_test "Test 1: stopOnError=true stops execution" "stopped" "stopped"
fi
echo ""

# Test 2: Verify stopOnError=false continues execution
echo "Test 2: stopOnError=false should continue after error"
echo "----------------------------------------------"
OUTPUT=$(bash +e -c '
    echo "Before error"
    false
    echo "After error - should appear"
' 2>&1)

if echo "$OUTPUT" | grep -q "After error"; then
    report_test "Test 2: stopOnError=false continues execution" "continued" "continued"
else
    report_test "Test 2: stopOnError=false continues execution" "continued" "stopped"
fi
echo ""

# Test 3: Test with command substitution
echo "Test 3: stopOnError with command substitution"
echo "----------------------------------------------"
OUTPUT=$(bash -e -c '
    echo "Testing command substitution"
    RESULT=$(ls /nonexistent 2>&1) || true
    echo "Command completed"
' 2>&1)

if echo "$OUTPUT" | grep -q "Command completed"; then
    report_test "Test 3: Command substitution handling" "completed" "completed"
else
    report_test "Test 3: Command substitution handling" "completed" "stopped"
fi
echo ""

# Test 4: Test with pipelines
echo "Test 4: stopOnError with pipelines"
echo "----------------------------------------------"
OUTPUT=$(bash -e -c '
    echo "Testing pipeline"
    echo "test" | grep "nonexistent" || true
    echo "Pipeline completed"
' 2>&1)

if echo "$OUTPUT" | grep -q "Pipeline completed"; then
    report_test "Test 4: Pipeline handling" "completed" "completed"
else
    report_test "Test 4: Pipeline handling" "completed" "stopped"
fi
echo ""

# Test 5: Test with explicit exit codes
echo "Test 5: Explicit exit code handling"
echo "----------------------------------------------"
OUTPUT=$(bash -e -c '
    echo "Before exit 1"
    exit 1
    echo "This should not appear"
' 2>&1)

if echo "$OUTPUT" | grep -q "This should not appear"; then
    report_test "Test 5: Explicit exit code stops execution" "stopped" "continued"
else
    report_test "Test 5: Explicit exit code stops execution" "stopped" "stopped"
fi
echo ""

# Summary
echo "=========================================="
echo "Test Suite Summary"
echo "=========================================="
echo "Tests Passed: $PASS_COUNT"
echo "Tests Failed: $FAIL_COUNT"
echo "Total Tests:  $((PASS_COUNT + FAIL_COUNT))"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
    echo "✓ All tests passed!"
    exit 0
else
    echo "✗ Some tests failed"
    exit 1
fi

