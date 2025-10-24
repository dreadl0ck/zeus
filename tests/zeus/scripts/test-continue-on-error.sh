#!/bin/bash
# Test script for stopOnError=false
# This script should continue execution even after failing commands

echo "Step 1: Starting test with stopOnError disabled"
echo "Step 2: This step will succeed"

# This command will fail (exit code 1)
echo "Step 3: About to execute a failing command..."
false

# These lines SHOULD execute because stopOnError is false
echo "Step 4: Continuing after error (stopOnError=false)"
echo "Step 5: Script completed successfully despite earlier failure"
exit 0

