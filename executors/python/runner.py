#!/usr/bin/env python3
"""
Python code executor for Scribble.
Reads user code and test cases from environment variables,
executes the code against test cases, and returns results as JSON.
"""

import os
import sys
import json
import time
import base64
import traceback
from io import StringIO
from typing import Any, Dict, List, Tuple, Optional

# Import psutil for memory measurement
try:
    import psutil
except ImportError:
    psutil = None


def measure_memory() -> int:
    """
    Measure current process memory usage in KB
    Returns RSS (Resident Set Size) for current memory usage
    """
    if psutil is None:
        return 0
    try:
        process = psutil.Process()
        memory_info = process.memory_info()
        return memory_info.rss // 1024  # Convert bytes to KB
    except Exception:
        return 0


def compare_outputs(actual: Any, expected: Any) -> bool:
    """
    Compare actual output with expected output
    Handles different data types and floating point comparisons
    """
    if actual is None and expected is None:
        return True
    if actual is None or expected is None:
        return False

    # For floating point numbers, use approximate comparison
    if isinstance(actual, float) and isinstance(expected, float):
        return abs(actual - expected) < 1e-9

    # For lists, compare recursively
    if isinstance(actual, list) and isinstance(expected, list):
        if len(actual) != len(expected):
            return False
        return all(compare_outputs(a, e) for a, e in zip(actual, expected))

    # Default: direct equality
    return actual == expected


def execute_test_case(code: str, test_input: Any) -> Tuple[Any, int, int, Optional[Dict]]:
    """
    Execute user code against a single test case

    Returns: (output, execution_time_ms, memory_kb, error_dict)
    """
    start_time = time.time()
    start_memory = measure_memory()

    # Capture stdout
    captured_output = StringIO()
    old_stdout = sys.stdout
    sys.stdout = captured_output

    try:
        # Create isolated namespace for code execution
        namespace = {}

        # Execute the user's code to define functions
        exec(code, namespace)

        # Find the solution function
        # Try common function names first
        func = None
        for func_name in ['solution', 'solve', 'main']:
            if func_name in namespace and callable(namespace[func_name]):
                func = namespace[func_name]
                break

        # If no standard function found, look for any callable
        if func is None:
            for name, obj in namespace.items():
                if callable(obj) and not name.startswith('_'):
                    func = obj
                    break

        if func is None:
            raise ValueError("No callable function found in submitted code")

        # Execute the function with test inputs
        if isinstance(test_input, dict):
            # Multiple parameters as kwargs
            result = func(**test_input)
        elif isinstance(test_input, list):
            # List as args
            result = func(*test_input)
        else:
            # Single parameter
            result = func(test_input)

        # Measure execution metrics
        end_time = time.time()
        end_memory = measure_memory()

        execution_time_ms = int((end_time - start_time) * 1000)
        memory_used_kb = max(end_memory - start_memory, 0)

        # Restore stdout
        sys.stdout = old_stdout

        return result, execution_time_ms, memory_used_kb, None

    except Exception as e:
        # Restore stdout
        sys.stdout = old_stdout

        # Capture error details
        error_trace = traceback.format_exc()
        end_time = time.time()
        execution_time_ms = int((end_time - start_time) * 1000)

        return None, execution_time_ms, 0, {
            "type": type(e).__name__,
            "message": str(e),
            "traceback": error_trace
        }


def main():
    """Main executor entry point."""

    try:
        # Get input from environment variables
        code_b64 = os.getenv('CODE', '')
        test_cases_json = os.getenv('TEST_CASES', '[]')
        problem_id = os.getenv('PROBLEM_ID', 'unknown')

        if not code_b64:
            print(json.dumps({
                'status': 'compilation_error',
                'error_message': 'No code provided',
                'execution_time_ms': 0,
                'memory_used_kb': 0,
                'tests_passed': 0,
                'tests_total': 0
            }))
            return

        # Decode code
        try:
            code = base64.b64decode(code_b64).decode('utf-8')
        except Exception as e:
            print(json.dumps({
                'status': 'compilation_error',
                'error_message': f'Failed to decode code: {str(e)}',
                'execution_time_ms': 0,
                'memory_used_kb': 0,
                'tests_passed': 0,
                'tests_total': 0
            }))
            return

        # Parse test cases
        try:
            test_cases = json.loads(test_cases_json)
            if not isinstance(test_cases, list):
                test_cases = [test_cases]
        except Exception as e:
            print(json.dumps({
                'status': 'runtime_error',
                'error_message': f'Failed to parse test cases: {str(e)}',
                'execution_time_ms': 0,
                'memory_used_kb': 0,
                'tests_passed': 0,
                'tests_total': 0
            }))
            return

        # Execute code against all test cases
        results = []
        total_tests = len(test_cases)
        passed_tests = 0
        total_execution_time = 0
        total_memory = 0

        for i, test_case in enumerate(test_cases):
            test_input = test_case.get('input')
            expected_output = test_case.get('expected_output')

            actual_output, exec_time, memory_used, error = execute_test_case(code, test_input)

            total_execution_time += exec_time
            total_memory += memory_used

            # Check if test passed
            passed = False
            if error is None:
                passed = compare_outputs(actual_output, expected_output)
                if passed:
                    passed_tests += 1

            results.append({
                "test_case_id": i,
                "input": test_input,
                "expected_output": expected_output,
                "actual_output": actual_output if error is None else None,
                "passed": passed,
                "execution_time_ms": exec_time,
                "memory_used_kb": memory_used,
                "error": error.get('message') if error else None
            })

        # Calculate averages
        avg_execution_time = total_execution_time // total_tests if total_tests > 0 else 0
        avg_memory = total_memory // total_tests if total_tests > 0 else 0

        # Determine overall status
        if passed_tests == total_tests:
            status = "accepted"
        elif any(r.get('error') for r in results):
            status = "runtime_error"
        else:
            status = "wrong_answer"

        # Output final results as JSON
        output = {
            "status": status,
            "tests_passed": passed_tests,
            "tests_total": total_tests,
            "execution_time_ms": avg_execution_time,
            "memory_used_kb": avg_memory,
            "test_results": results
        }

        print(json.dumps(output, indent=2))

    except Exception as e:
        # Unexpected error in executor itself
        print(json.dumps({
            'status': 'runtime_error',
            'error_message': f'Executor error: {str(e)}',
            'execution_time_ms': 0,
            'memory_used_kb': 0,
            'tests_passed': 0,
            'tests_total': 0,
            'traceback': traceback.format_exc()
        }))


if __name__ == '__main__':
    main()
