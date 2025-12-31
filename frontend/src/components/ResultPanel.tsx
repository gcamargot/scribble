import { useState } from 'react';

interface TestCaseResult {
  id: number;
  input: string;
  expected: string;
  actual: string;
  passed: boolean;
  executionTime: number; // in ms
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
          <h3
            className={`font-bold text-xl ${
              isAccepted ? 'text-green-300' : 'text-red-300'
            }`}
          >
            {result.verdict}
          </h3>

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
