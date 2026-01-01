import { useState } from 'react';

interface TestCaseResult {
  id: number;
  input: string;
  expected: string;
  actual: string;
  passed: boolean;
  executionTime: number; // in ms
}

/**
 * Error types that can occur during code execution
 */
export type ErrorType =
  | 'compilation_error'
  | 'time_limit_exceeded'
  | 'memory_limit_exceeded'
  | 'runtime_error'
  | 'wrong_answer'
  | 'internal_error'
  | null;

/**
 * Extended error details for debugging
 */
export interface ErrorDetails {
  type: ErrorType;
  line?: number;
  column?: number;
  compilerOutput?: string;
  stackTrace?: string;
  exitCode?: number;
  signal?: string;
  hint?: string;
}

export interface SubmissionResult {
  success: boolean;
  status: string;
  verdict: string;
  executionTime: string;
  executionTimeMs?: number;
  memoryUsed: string;
  memoryUsedMb?: number;
  testsPassed: number;
  testsTotal: number;
  percentile: {
    time: number;
    memory: number;
  };
  message: string;
  testCases?: TestCaseResult[];
  errorDetails?: ErrorDetails;
}

interface ResultPanelProps {
  result: SubmissionResult;
}

/**
 * Get color class based on execution time
 * Green: <100ms, Yellow: <1s, Red: >1s
 */
function getTimeColor(timeMs: number): string {
  if (timeMs < 100) return 'bg-green-500';
  if (timeMs < 1000) return 'bg-yellow-500';
  return 'bg-red-500';
}

/**
 * Get text color class based on execution time
 */
function getTimeTextColor(timeMs: number): string {
  if (timeMs < 100) return 'text-green-400';
  if (timeMs < 1000) return 'text-yellow-400';
  return 'text-red-400';
}

/**
 * Parse execution time string to milliseconds
 */
function parseTimeToMs(timeStr: string): number {
  const match = timeStr.match(/(\d+(?:\.\d+)?)\s*(ms|s|m)/i);
  if (!match) return 0;

  const value = parseFloat(match[1]);
  const unit = match[2].toLowerCase();

  switch (unit) {
    case 'ms':
      return value;
    case 's':
      return value * 1000;
    case 'm':
      return value * 60000;
    default:
      return value;
  }
}

/**
 * Parse memory string to MB
 */
function parseMemoryToMb(memStr: string): number {
  const match = memStr.match(/(\d+(?:\.\d+)?)\s*(kb|mb|gb)/i);
  if (!match) return 0;

  const value = parseFloat(match[1]);
  const unit = match[2].toLowerCase();

  switch (unit) {
    case 'kb':
      return value / 1024;
    case 'mb':
      return value;
    case 'gb':
      return value * 1024;
    default:
      return value;
  }
}

/**
 * Detect error type from verdict string
 */
function detectErrorType(verdict: string): ErrorType {
  const v = verdict.toLowerCase();
  if (v.includes('compilation') || v.includes('compile')) return 'compilation_error';
  if (v.includes('time') && v.includes('limit')) return 'time_limit_exceeded';
  if (v.includes('memory') && v.includes('limit')) return 'memory_limit_exceeded';
  if (v.includes('runtime') || v.includes('error')) return 'runtime_error';
  if (v.includes('wrong')) return 'wrong_answer';
  if (v.includes('internal')) return 'internal_error';
  if (v !== 'accepted') return 'wrong_answer';
  return null;
}

/**
 * Get user-friendly error message based on error type
 */
function getErrorMessage(errorType: ErrorType): { title: string; description: string; icon: string } {
  switch (errorType) {
    case 'compilation_error':
      return {
        title: 'Compilation Error',
        description: 'Your code failed to compile. Check for syntax errors, missing imports, or type mismatches.',
        icon: '🔧'
      };
    case 'time_limit_exceeded':
      return {
        title: 'Time Limit Exceeded',
        description: 'Your solution took too long to execute. Consider optimizing your algorithm or reducing time complexity.',
        icon: '⏱️'
      };
    case 'memory_limit_exceeded':
      return {
        title: 'Memory Limit Exceeded',
        description: 'Your solution used too much memory. Consider using more memory-efficient data structures.',
        icon: '💾'
      };
    case 'runtime_error':
      return {
        title: 'Runtime Error',
        description: 'Your code crashed during execution. Check for division by zero, array out of bounds, or null pointer errors.',
        icon: '💥'
      };
    case 'wrong_answer':
      return {
        title: 'Wrong Answer',
        description: 'Your solution produced incorrect output for one or more test cases.',
        icon: '❌'
      };
    case 'internal_error':
      return {
        title: 'Internal Error',
        description: 'An unexpected error occurred on our end. Please try again or report this issue.',
        icon: '⚠️'
      };
    default:
      return {
        title: 'Error',
        description: 'An error occurred during execution.',
        icon: '❓'
      };
  }
}

/**
 * Error details expandable section
 */
function ErrorDetailsSection({ details }: { details: ErrorDetails }) {
  const [isExpanded, setIsExpanded] = useState(false);
  const hasDetails = details.compilerOutput || details.stackTrace || details.hint;

  if (!hasDetails) return null;

  return (
    <div className="mt-4">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-center justify-between p-3 bg-gray-800 hover:bg-gray-700 rounded-lg transition-colors border border-gray-600"
      >
        <span className="text-gray-300 font-medium text-sm">
          Error Details
        </span>
        <span className="text-gray-400">
          {isExpanded ? '▼ Hide' : '▶ Show'}
        </span>
      </button>

      {isExpanded && (
        <div className="mt-2 bg-gray-900 rounded-lg p-4 space-y-4 border border-gray-700">
          {/* Location info */}
          {(details.line || details.column) && (
            <div>
              <span className="text-gray-400 text-xs font-medium block mb-1">Location:</span>
              <span className="text-yellow-400 font-mono text-sm">
                Line {details.line}{details.column ? `, Column ${details.column}` : ''}
              </span>
            </div>
          )}

          {/* Compiler output */}
          {details.compilerOutput && (
            <div>
              <span className="text-gray-400 text-xs font-medium block mb-1">Compiler Output:</span>
              <pre className="text-red-400 font-mono text-xs bg-gray-800 p-3 rounded overflow-x-auto max-h-48 overflow-y-auto whitespace-pre-wrap">
                {details.compilerOutput}
              </pre>
            </div>
          )}

          {/* Stack trace */}
          {details.stackTrace && (
            <div>
              <span className="text-gray-400 text-xs font-medium block mb-1">Stack Trace:</span>
              <pre className="text-orange-400 font-mono text-xs bg-gray-800 p-3 rounded overflow-x-auto max-h-48 overflow-y-auto whitespace-pre-wrap">
                {details.stackTrace}
              </pre>
            </div>
          )}

          {/* Exit code / signal */}
          {(details.exitCode !== undefined || details.signal) && (
            <div className="flex gap-4">
              {details.exitCode !== undefined && (
                <div>
                  <span className="text-gray-400 text-xs font-medium block mb-1">Exit Code:</span>
                  <span className="text-red-400 font-mono text-sm">{details.exitCode}</span>
                </div>
              )}
              {details.signal && (
                <div>
                  <span className="text-gray-400 text-xs font-medium block mb-1">Signal:</span>
                  <span className="text-red-400 font-mono text-sm">{details.signal}</span>
                </div>
              )}
            </div>
          )}

          {/* Hint */}
          {details.hint && (
            <div className="bg-blue-900/30 border border-blue-700 rounded-lg p-3">
              <span className="text-blue-400 text-xs font-medium block mb-1">💡 Hint:</span>
              <p className="text-blue-300 text-sm">{details.hint}</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

/**
 * Progress bar component
 */
function ProgressBar({
  value,
  max,
  colorClass,
  showPercentage = true
}: {
  value: number;
  max: number;
  colorClass: string;
  showPercentage?: boolean;
}) {
  const percentage = Math.min((value / max) * 100, 100);

  return (
    <div className="w-full bg-gray-700 rounded-full h-2.5 overflow-hidden">
      <div
        className={`h-full rounded-full transition-all duration-500 ${colorClass}`}
        style={{ width: `${percentage}%` }}
      />
      {showPercentage && (
        <span className="sr-only">{Math.round(percentage)}%</span>
      )}
    </div>
  );
}

/**
 * Test case result row component
 */
function TestCaseRow({ testCase, index }: { testCase: TestCaseResult; index: number }) {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div className="border border-gray-700 rounded-lg overflow-hidden">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className={`w-full flex items-center justify-between p-3 text-left transition-colors ${
          testCase.passed
            ? 'bg-green-900/30 hover:bg-green-900/50'
            : 'bg-red-900/30 hover:bg-red-900/50'
        }`}
      >
        <div className="flex items-center gap-3">
          <span
            className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ${
              testCase.passed
                ? 'bg-green-600 text-white'
                : 'bg-red-600 text-white'
            }`}
          >
            {testCase.passed ? '✓' : '✗'}
          </span>
          <span className="text-gray-300 text-sm font-medium">
            Test Case {index + 1}
          </span>
          <span className="text-gray-500 text-xs">
            {testCase.executionTime}ms
          </span>
        </div>
        <span className="text-gray-400 text-sm">
          {isExpanded ? '▼' : '▶'}
        </span>
      </button>

      {isExpanded && (
        <div className="bg-gray-900 p-4 space-y-3 border-t border-gray-700">
          <div>
            <span className="text-gray-400 text-xs font-medium block mb-1">Input:</span>
            <pre className="text-blue-400 font-mono text-sm bg-gray-800 p-2 rounded overflow-x-auto">
              {testCase.input}
            </pre>
          </div>
          <div>
            <span className="text-gray-400 text-xs font-medium block mb-1">Expected:</span>
            <pre className="text-green-400 font-mono text-sm bg-gray-800 p-2 rounded overflow-x-auto">
              {testCase.expected}
            </pre>
          </div>
          {!testCase.passed && (
            <div>
              <span className="text-gray-400 text-xs font-medium block mb-1">Your Output:</span>
              <pre className="text-red-400 font-mono text-sm bg-gray-800 p-2 rounded overflow-x-auto">
                {testCase.actual}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

/**
 * ResultPanel - Enhanced results display with metrics
 */
export default function ResultPanel({ result }: ResultPanelProps) {
  const [showTestCases, setShowTestCases] = useState(false);

  // Parse metrics
  const timeMs = result.executionTimeMs ?? parseTimeToMs(result.executionTime);
  const memoryMb = result.memoryUsedMb ?? parseMemoryToMb(result.memoryUsed);

  // Calculate test pass rate
  const passRate = result.testsTotal > 0
    ? Math.round((result.testsPassed / result.testsTotal) * 100)
    : 0;

  const isAccepted = result.verdict === 'Accepted';

  // Detect error type and get user-friendly message
  const errorType = result.errorDetails?.type ?? detectErrorType(result.verdict);
  const errorInfo = errorType ? getErrorMessage(errorType) : null;

  // Mock test cases if not provided (for backward compatibility)
  const testCases = result.testCases || [];

  return (
    <div className="bg-gray-800 border-t border-gray-700 p-4 space-y-4">
      {/* Verdict Header */}
      <div
        className={`rounded-lg p-4 ${
          isAccepted
            ? 'bg-green-900/50 border border-green-700'
            : 'bg-red-900/50 border border-red-700'
        }`}
      >
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            {errorInfo && (
              <span className="text-2xl" role="img" aria-label={errorInfo.title}>
                {errorInfo.icon}
              </span>
            )}
            <h3
              className={`font-bold text-xl ${
                isAccepted ? 'text-green-300' : 'text-red-300'
              }`}
            >
              {result.verdict}
            </h3>
          </div>

          {/* Test Pass Rate Badge */}
          <div className="flex items-center gap-2">
            <div
              className={`px-3 py-1 rounded-full text-sm font-semibold ${
                passRate === 100
                  ? 'bg-green-600 text-white'
                  : passRate >= 50
                    ? 'bg-yellow-600 text-white'
                    : 'bg-red-600 text-white'
              }`}
            >
              {passRate}% Passed
            </div>
            <span className="text-gray-400 text-sm">
              ({result.testsPassed}/{result.testsTotal})
            </span>
          </div>
        </div>

        {/* Metrics Grid */}
        <div className="grid grid-cols-2 gap-4">
          {/* Execution Time */}
          <div className="bg-gray-900 rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-gray-400 text-sm font-medium">Execution Time</span>
              <span className={`font-mono font-bold ${getTimeTextColor(timeMs)}`}>
                {result.executionTime}
              </span>
            </div>
            <ProgressBar
              value={timeMs}
              max={2000} // 2 seconds max for visualization
              colorClass={getTimeColor(timeMs)}
            />
            <p className="text-gray-500 text-xs mt-2">
              Faster than {result.percentile.time}% of submissions
            </p>
          </div>

          {/* Memory Usage */}
          <div className="bg-gray-900 rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-gray-400 text-sm font-medium">Memory Used</span>
              <span className="text-blue-400 font-mono font-bold">
                {result.memoryUsed}
              </span>
            </div>
            <ProgressBar
              value={memoryMb}
              max={512} // 512MB max for visualization
              colorClass="bg-blue-500"
            />
            <p className="text-gray-500 text-xs mt-2">
              Better than {result.percentile.memory}% of submissions
            </p>
          </div>
        </div>

        {/* Message */}
        <p className="text-gray-300 text-sm mt-4">{result.message}</p>

        {/* Error Description */}
        {errorInfo && !isAccepted && (
          <div className="mt-3 p-3 bg-gray-800 rounded-lg border border-gray-600">
            <p className="text-gray-400 text-sm">{errorInfo.description}</p>
          </div>
        )}

        {/* Error Details (expandable) */}
        {result.errorDetails && <ErrorDetailsSection details={result.errorDetails} />}
      </div>

      {/* Expandable Test Cases */}
      {testCases.length > 0 && (
        <div>
          <button
            onClick={() => setShowTestCases(!showTestCases)}
            className="w-full flex items-center justify-between p-3 bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors"
          >
            <span className="text-gray-300 font-medium">
              Test Case Details
            </span>
            <span className="text-gray-400">
              {showTestCases ? '▼ Hide' : '▶ Show'}
            </span>
          </button>

          {showTestCases && (
            <div className="mt-3 space-y-2">
              {testCases.map((tc, index) => (
                <TestCaseRow key={tc.id || index} testCase={tc} index={index} />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
