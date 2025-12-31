#!/usr/bin/env node
/**
 * JavaScript code executor for Scribble.
 * Reads user code and test cases from environment variables,
 * executes the code against test cases, and returns results as JSON.
 *
 * Memory measurement: Uses process.memoryUsage() for heap tracking.
 * Time tracking: Separates compilation time from execution time per test.
 */

const fs = require('fs');
const vm = require('vm');
const path = require('path');

/**
 * Read cgroups v2 memory peak (container environments)
 */
function getCgroupsMemoryPeak() {
    try {
        const data = fs.readFileSync('/sys/fs/cgroup/memory.peak', 'utf8');
        return Math.floor(parseInt(data.trim(), 10) / 1024); // bytes to KB
    } catch {
        return null;
    }
}

/**
 * Get current memory usage in KB
 */
function measureMemory() {
    // Try cgroups first
    const cgroupsMem = getCgroupsMemoryPeak();
    if (cgroupsMem !== null) {
        return cgroupsMem;
    }

    // Fallback to Node.js memory usage
    const usage = process.memoryUsage();
    return Math.floor(usage.heapUsed / 1024);
}

/**
 * Memory monitor that samples every 10ms
 */
class MemoryMonitor {
    constructor(intervalMs = 10) {
        this.intervalMs = intervalMs;
        this.peakMemoryKb = 0;
        this.baseline = 0;
        this.timer = null;
    }

    start() {
        this.baseline = measureMemory();
        this.peakMemoryKb = this.baseline;
        this.timer = setInterval(() => {
            const current = measureMemory();
            if (current > this.peakMemoryKb) {
                this.peakMemoryKb = current;
            }
        }, this.intervalMs);
    }

    stop() {
        if (this.timer) {
            clearInterval(this.timer);
            this.timer = null;
        }
        return Math.max(this.peakMemoryKb - this.baseline, 0);
    }
}

/**
 * Check if OOM kill occurred
 */
function checkOomKilled() {
    try {
        const data = fs.readFileSync('/sys/fs/cgroup/memory.events', 'utf8');
        const lines = data.split('\n');
        for (const line of lines) {
            if (line.startsWith('oom_kill')) {
                const parts = line.split(/\s+/);
                if (parts.length >= 2 && parseInt(parts[1], 10) > 0) {
                    return true;
                }
            }
        }
    } catch {
        // Not in container environment
    }
    return false;
}

/**
 * Compare outputs with floating point tolerance
 */
function compareOutputs(actual, expected) {
    if (actual === null && expected === null) return true;
    if (actual === null || expected === null) return false;

    if (typeof actual === 'number' && typeof expected === 'number') {
        return Math.abs(actual - expected) < 1e-9;
    }

    if (Array.isArray(actual) && Array.isArray(expected)) {
        if (actual.length !== expected.length) return false;
        return actual.every((a, i) => compareOutputs(a, expected[i]));
    }

    if (typeof actual === 'object' && typeof expected === 'object') {
        const aKeys = Object.keys(actual);
        const eKeys = Object.keys(expected);
        if (aKeys.length !== eKeys.length) return false;
        return aKeys.every(key => compareOutputs(actual[key], expected[key]));
    }

    return actual === expected;
}

/**
 * Compile user code and find the solution function
 */
function compileCode(code) {
    const startTime = process.hrtime.bigint();

    try {
        // Create a sandbox context
        const sandbox = {
            console: { log: () => {}, error: () => {}, warn: () => {} },
            Math,
            Array,
            Object,
            String,
            Number,
            Boolean,
            JSON,
            Map,
            Set,
            Date,
            RegExp,
            Error,
            parseInt,
            parseFloat,
            isNaN,
            isFinite,
        };

        const context = vm.createContext(sandbox);

        // Execute user code to define functions
        const script = new vm.Script(code, { timeout: 5000 });
        script.runInContext(context);

        // Find the solution function
        const funcNames = ['solution', 'solve', 'main', 'Solution'];
        let func = null;

        for (const name of funcNames) {
            if (typeof sandbox[name] === 'function') {
                func = sandbox[name];
                break;
            }
        }

        // Look for any exported function
        if (!func) {
            for (const key of Object.keys(sandbox)) {
                if (typeof sandbox[key] === 'function' && !key.startsWith('_')) {
                    func = sandbox[key];
                    break;
                }
            }
        }

        const endTime = process.hrtime.bigint();
        const compilationTimeMs = Number(endTime - startTime) / 1e6;

        if (!func) {
            return {
                func: null,
                compilationTimeMs: Math.floor(compilationTimeMs),
                error: {
                    type: 'CompilationError',
                    message: 'No callable function found in submitted code'
                }
            };
        }

        return { func, context, compilationTimeMs: Math.floor(compilationTimeMs), error: null };
    } catch (e) {
        const endTime = process.hrtime.bigint();
        const compilationTimeMs = Number(endTime - startTime) / 1e6;
        return {
            func: null,
            compilationTimeMs: Math.floor(compilationTimeMs),
            error: {
                type: e.name || 'SyntaxError',
                message: e.message,
                traceback: e.stack
            }
        };
    }
}

/**
 * Execute function against a single test case
 */
function executeTestCase(func, testInput) {
    const startTime = process.hrtime.bigint();

    try {
        let result;
        if (typeof testInput === 'object' && !Array.isArray(testInput) && testInput !== null) {
            // Object with named parameters
            result = func(testInput);
        } else if (Array.isArray(testInput)) {
            result = func(...testInput);
        } else {
            result = func(testInput);
        }

        const endTime = process.hrtime.bigint();
        const executionTimeMs = Number(endTime - startTime) / 1e6;

        if (checkOomKilled()) {
            return {
                output: null,
                executionTimeMs: Math.floor(executionTimeMs),
                error: { type: 'MemoryError', message: 'Execution exceeded memory limit (OOM killed)' }
            };
        }

        return { output: result, executionTimeMs: Math.floor(executionTimeMs), error: null };
    } catch (e) {
        const endTime = process.hrtime.bigint();
        const executionTimeMs = Number(endTime - startTime) / 1e6;
        return {
            output: null,
            executionTimeMs: Math.floor(executionTimeMs),
            error: {
                type: e.name || 'RuntimeError',
                message: e.message,
                traceback: e.stack
            }
        };
    }
}

/**
 * Main executor entry point
 */
function main() {
    try {
        const codeB64 = process.env.CODE || '';
        const testCasesJson = process.env.TEST_CASES || '[]';
        const problemId = process.env.PROBLEM_ID || 'unknown';

        if (!codeB64) {
            console.log(JSON.stringify({
                status: 'compilation_error',
                error_message: 'No code provided',
                compilation_time_ms: 0,
                execution_time_ms: 0,
                memory_used_kb: 0,
                tests_passed: 0,
                tests_total: 0
            }));
            return;
        }

        // Decode code
        let code;
        try {
            code = Buffer.from(codeB64, 'base64').toString('utf8');
        } catch (e) {
            console.log(JSON.stringify({
                status: 'compilation_error',
                error_message: `Failed to decode code: ${e.message}`,
                compilation_time_ms: 0,
                execution_time_ms: 0,
                memory_used_kb: 0,
                tests_passed: 0,
                tests_total: 0
            }));
            return;
        }

        // Parse test cases
        let testCases;
        try {
            testCases = JSON.parse(testCasesJson);
            if (!Array.isArray(testCases)) {
                testCases = [testCases];
            }
        } catch (e) {
            console.log(JSON.stringify({
                status: 'runtime_error',
                error_message: `Failed to parse test cases: ${e.message}`,
                compilation_time_ms: 0,
                execution_time_ms: 0,
                memory_used_kb: 0,
                tests_passed: 0,
                tests_total: 0
            }));
            return;
        }

        // Start memory monitoring
        const memoryMonitor = new MemoryMonitor(10);
        memoryMonitor.start();

        // Compile code
        const { func, compilationTimeMs, error: compileError } = compileCode(code);

        if (compileError) {
            memoryMonitor.stop();
            console.log(JSON.stringify({
                status: 'compilation_error',
                error_message: compileError.message,
                error_type: compileError.type,
                compilation_time_ms: compilationTimeMs,
                execution_time_ms: 0,
                memory_used_kb: 0,
                tests_passed: 0,
                tests_total: 0
            }));
            return;
        }

        // Execute tests
        const results = [];
        let passedTests = 0;
        let totalExecutionTime = 0;
        let oomDetected = false;

        for (let i = 0; i < testCases.length; i++) {
            const testCase = testCases[i];
            const testInput = testCase.input;
            const expectedOutput = testCase.expected_output;

            const { output, executionTimeMs, error } = executeTestCase(func, testInput);
            totalExecutionTime += executionTimeMs;

            if (error && error.type === 'MemoryError') {
                oomDetected = true;
            }

            let passed = false;
            if (!error) {
                passed = compareOutputs(output, expectedOutput);
                if (passed) passedTests++;
            }

            results.push({
                test_case_id: i,
                passed,
                actual_output: error ? null : output,
                execution_time_ms: executionTimeMs,
                error: error ? error.message : null
            });
        }

        // Stop memory monitoring
        let peakMemoryKb = memoryMonitor.stop();

        // Check cgroups for more accurate reading
        const cgroupsPeak = getCgroupsMemoryPeak();
        if (cgroupsPeak !== null) {
            peakMemoryKb = cgroupsPeak;
        }

        const avgExecutionTime = testCases.length > 0 ? Math.floor(totalExecutionTime / testCases.length) : 0;

        // Determine status
        let status;
        if (oomDetected) {
            status = 'memory_limit';
        } else if (passedTests === testCases.length) {
            status = 'accepted';
        } else if (results.some(r => r.error)) {
            status = 'runtime_error';
        } else {
            status = 'wrong_answer';
        }

        console.log(JSON.stringify({
            status,
            tests_passed: passedTests,
            tests_total: testCases.length,
            compilation_time_ms: compilationTimeMs,
            execution_time_ms: avgExecutionTime,
            total_execution_time_ms: Math.floor(totalExecutionTime),
            memory_used_kb: peakMemoryKb,
            test_results: results
        }));

    } catch (e) {
        console.log(JSON.stringify({
            status: 'runtime_error',
            error_message: `Executor error: ${e.message}`,
            compilation_time_ms: 0,
            execution_time_ms: 0,
            memory_used_kb: 0,
            tests_passed: 0,
            tests_total: 0,
            traceback: e.stack
        }));
    }
}

main();
