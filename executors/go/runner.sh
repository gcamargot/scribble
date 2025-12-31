#!/bin/sh
# Go code executor for Scribble
# Compiles and executes user Go code against test cases

set -e

# Helper function to output JSON error
print_error() {
    local status="$1"
    local message="$2"
    local compile_time="${3:-0}"
    echo "{\"status\":\"$status\",\"error_message\":\"$(echo "$message" | sed 's/"/\\"/g' | tr '\n' ' ')\",\"compilation_time_ms\":$compile_time,\"execution_time_ms\":0,\"memory_used_kb\":0,\"tests_passed\":0,\"tests_total\":0}"
    exit 0
}

# Check for code
if [ -z "$CODE" ]; then
    print_error "compilation_error" "No code provided"
fi

# Decode base64 code
echo "$CODE" | base64 -d > /tmp/solution.go 2>/dev/null || print_error "compilation_error" "Failed to decode code"

# Compile with timing
compile_start=$(date +%s%3N)
compile_output=$(go build -o /tmp/solution /tmp/solution.go 2>&1) || {
    compile_end=$(date +%s%3N)
    compile_time=$((compile_end - compile_start))
    print_error "compilation_error" "$compile_output" "$compile_time"
}
compile_end=$(date +%s%3N)
compile_time=$((compile_end - compile_start))

# Parse test cases and execute using Python
python3 << PYTHON_RUNNER
import json
import subprocess
import os
import time
import sys

test_cases_json = os.getenv('TEST_CASES', '[]')
compile_time = $compile_time

try:
    test_cases = json.loads(test_cases_json)
    if not isinstance(test_cases, list):
        test_cases = [test_cases]
except:
    print(json.dumps({
        'status': 'runtime_error',
        'error_message': 'Failed to parse test cases',
        'compilation_time_ms': compile_time,
        'execution_time_ms': 0,
        'memory_used_kb': 0,
        'tests_passed': 0,
        'tests_total': 0
    }))
    sys.exit(0)

results = []
passed = 0
total_exec_time = 0

for i, tc in enumerate(test_cases):
    test_input = tc.get('input', '')
    expected = tc.get('expected_output')

    # Convert input to string for stdin
    if isinstance(test_input, list):
        stdin_data = ' '.join(str(x) for x in test_input)
    else:
        stdin_data = str(test_input)

    start = time.time()
    try:
        result = subprocess.run(
            ['/tmp/solution'],
            input=stdin_data,
            capture_output=True,
            text=True,
            timeout=5
        )
        output = result.stdout.strip()
        error = result.stderr if result.returncode != 0 else None
    except subprocess.TimeoutExpired:
        output = None
        error = 'Time limit exceeded'
    except Exception as e:
        output = None
        error = str(e)

    end = time.time()
    exec_time = int((end - start) * 1000)
    total_exec_time += exec_time

    # Try to parse output
    actual = None
    if output and not error:
        try:
            actual = json.loads(output)
        except:
            try:
                actual = int(output)
            except:
                try:
                    actual = float(output)
                except:
                    actual = output

    # Compare
    test_passed = False
    if error is None and actual is not None:
        if isinstance(expected, float) and isinstance(actual, (int, float)):
            test_passed = abs(float(actual) - expected) < 1e-9
        elif isinstance(expected, list) and isinstance(actual, list):
            test_passed = actual == expected
        else:
            test_passed = actual == expected

    if test_passed:
        passed += 1

    results.append({
        'test_case_id': i,
        'passed': test_passed,
        'actual_output': actual,
        'execution_time_ms': exec_time,
        'error': error
    })

# Get memory from cgroups
memory_kb = 0
try:
    with open('/sys/fs/cgroup/memory.peak', 'r') as f:
        memory_kb = int(f.read().strip()) // 1024
except:
    pass

avg_exec = total_exec_time // len(test_cases) if test_cases else 0

if passed == len(test_cases):
    status = 'accepted'
elif any(r.get('error') for r in results):
    status = 'runtime_error'
else:
    status = 'wrong_answer'

print(json.dumps({
    'status': status,
    'tests_passed': passed,
    'tests_total': len(test_cases),
    'compilation_time_ms': compile_time,
    'execution_time_ms': avg_exec,
    'total_execution_time_ms': total_exec_time,
    'memory_used_kb': memory_kb,
    'test_results': results
}))
PYTHON_RUNNER
