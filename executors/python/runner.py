#!/usr/bin/env python3
"""
Python code executor for Scribble.
Reads user code and test cases from environment variables,
executes the code against test cases, and returns results as JSON.

Memory measurement: Uses cgroups v2 memory.peak (preferred) or VmPeak sampling.
"""

import os
import sys
import json
import time
import base64
import traceback
import threading
from io import StringIO
from typing import Any, Dict, List, Tuple, Optional

# Import psutil for memory measurement fallback
try:
    import psutil
except ImportError:
    psutil = None


def get_cgroups_v2_memory_peak() -> Optional[int]:
    """
    Read peak memory from cgroups v2 (if available in container)
    Returns memory in KB, or None if not available
    """
    cgroup_memory_peak = "/sys/fs/cgroup/memory.peak"
    try:
        if os.path.exists(cgroup_memory_peak):
            with open(cgroup_memory_peak, 'r') as f:
                return int(f.read().strip()) // 1024  # bytes to KB
    except Exception:
        pass
    return None


def get_proc_vmpeak() -> Optional[int]:
    """
    Read VmPeak from /proc/self/status
    Returns memory in KB, or None if not available
    """
    try:
        with open('/proc/self/status', 'r') as f:
            for line in f:
                if line.startswith('VmPeak:'):
                    # Format: "VmPeak:    12345 kB"
                    parts = line.split()
                    if len(parts) >= 2:
                        return int(parts[1])  # Already in KB
    except Exception:
        pass
    return None


def get_proc_vmrss() -> Optional[int]:
    """
    Read current VmRSS from /proc/self/status
    Returns memory in KB, or None if not available
    """
    try:
        with open('/proc/self/status', 'r') as f:
            for line in f:
                if line.startswith('VmRSS:'):
                    parts = line.split()
                    if len(parts) >= 2:
                        return int(parts[1])
    except Exception:
        pass
    return None


def measure_memory() -> int:
    """
    Measure current process memory usage in KB
    Priority: cgroups v2 > VmRSS > psutil RSS
    """
    # Try cgroups v2 first (most accurate in containerized environments)
    cgroups_mem = get_cgroups_v2_memory_peak()
    if cgroups_mem is not None:
        return cgroups_mem

    # Try VmRSS from /proc/self/status
    vmrss = get_proc_vmrss()
    if vmrss is not None:
        return vmrss

    # Fallback to psutil RSS
    if psutil is None:
        return 0
    try:
        process = psutil.Process()
        memory_info = process.memory_info()
        return memory_info.rss // 1024  # Convert bytes to KB
    except Exception:
        return 0


class MemoryMonitor:
    """
    Background memory monitor that samples memory usage every 10ms
    to capture peak usage during execution
    """
    def __init__(self, interval_ms: int = 10):
        self.interval_ms = interval_ms
        self.peak_memory_kb = 0
        self._stop = False
        self._thread = None
        self._baseline = 0

    def _sample_memory(self):
        """Sample memory in a loop until stopped"""
        while not self._stop:
            current = measure_memory()
            if current > self.peak_memory_kb:
                self.peak_memory_kb = current
            time.sleep(self.interval_ms / 1000.0)

    def start(self):
        """Start the memory monitoring thread"""
        self._baseline = measure_memory()
        self._stop = False
        self._thread = threading.Thread(target=self._sample_memory, daemon=True)
        self._thread.start()

    def stop(self) -> int:
        """Stop monitoring and return peak memory usage in KB (above baseline)"""
        self._stop = True
        if self._thread:
            self._thread.join(timeout=0.1)
        # Return peak minus baseline to get actual usage during execution
        return max(self.peak_memory_kb - self._baseline, 0)


def check_oom_killed() -> bool:
    """
    Check if OOM kill occurred by reading cgroups memory events
    Returns True if OOM kill detected
    """
    try:
        with open('/sys/fs/cgroup/memory.events', 'r') as f:
            for line in f:
                if line.startswith('oom_kill'):
                    parts = line.split()
                    if len(parts) >= 2 and int(parts[1]) > 0:
                        return True
    except Exception:
        pass
    return False


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


def execute_test_case(code: str, test_input: Any, memory_monitor: MemoryMonitor) -> Tuple[Any, int, int, Optional[Dict]]:
    """
    Execute user code against a single test case

    Returns: (output, execution_time_ms, memory_kb, error_dict)
    """
    start_time = time.time()

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
        func = None
        for func_name in ['solution', 'solve', 'main', 'Solution']:
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
            result = func(**test_input)
        elif isinstance(test_input, list):
            result = func(*test_input)
        else:
            result = func(test_input)

        # Measure execution time
        end_time = time.time()
        execution_time_ms = int((end_time - start_time) * 1000)

        # Check for OOM kill
        if check_oom_killed():
            sys.stdout = old_stdout
            return None, execution_time_ms, 0, {
                "type": "MemoryError",
                "message": "Execution exceeded memory limit (OOM killed)"
            }

        # Restore stdout
        sys.stdout = old_stdout

        return result, execution_time_ms, 0, None

    except MemoryError as e:
        sys.stdout = old_stdout
        end_time = time.time()
        execution_time_ms = int((end_time - start_time) * 1000)
        return None, execution_time_ms, 0, {
            "type": "MemoryError",
            "message": "Execution exceeded memory limit"
        }

    except Exception as e:
        sys.stdout = old_stdout
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

        # Start memory monitoring (samples every 10ms)
        memory_monitor = MemoryMonitor(interval_ms=10)
        memory_monitor.start()

        # Execute code against all test cases
        results = []
        total_tests = len(test_cases)
        passed_tests = 0
        total_execution_time = 0
        oom_detected = False

        for i, test_case in enumerate(test_cases):
            test_input = test_case.get('input')
            expected_output = test_case.get('expected_output')

            actual_output, exec_time, _, error = execute_test_case(code, test_input, memory_monitor)

            total_execution_time += exec_time

            # Check for OOM
            if error and error.get('type') == 'MemoryError':
                oom_detected = True

            # Check if test passed
            passed = False
            if error is None:
                passed = compare_outputs(actual_output, expected_output)
                if passed:
                    passed_tests += 1

            results.append({
                "test_case_id": i,
                "passed": passed,
                "actual_output": actual_output if error is None else None,
                "execution_time_ms": exec_time,
                "error": error.get('message') if error else None
            })

        # Stop memory monitoring and get peak usage
        peak_memory_kb = memory_monitor.stop()

        # Also check cgroups v2 peak for most accurate reading
        cgroups_peak = get_cgroups_v2_memory_peak()
        if cgroups_peak is not None:
            peak_memory_kb = cgroups_peak

        # Also get VmPeak as another option
        vmpeak = get_proc_vmpeak()
        if vmpeak is not None and vmpeak > peak_memory_kb:
            peak_memory_kb = vmpeak

        # Calculate averages
        avg_execution_time = total_execution_time // total_tests if total_tests > 0 else 0

        # Determine overall status
        if oom_detected:
            status = "memory_limit"
        elif passed_tests == total_tests:
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
            "memory_used_kb": peak_memory_kb,
            "test_results": results
        }

        print(json.dumps(output))

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
