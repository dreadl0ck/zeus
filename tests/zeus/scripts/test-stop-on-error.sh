#!/bin/bash
# Test script for stopOnError=true
# This script should stop execution after the first failing command

echo "Step 1: Starting test with stopOnError enabled"
echo "Step 2: This step will succeed"

# This command will fail (exit code 1)
echo "Step 3: About to execute a failing command..."
false

# These lines should NOT execute because stopOnError is true
echo "ERROR: This line should NOT be printed!"
echo "ERROR: If you see this, stopOnError is not working correctly!"
exit 0

