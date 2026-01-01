import { useState } from 'react';
import type { SubmissionHistoryEntry } from '../hooks/useSubmissionHistory';
import { LANGUAGE_LABELS, type Language } from '../constants/starterCode';

interface SubmissionCardProps {
  submission: SubmissionHistoryEntry;
}

/**
 * Get verdict color based on status
 */
function getVerdictColor(verdict: string): string {
  const v = verdict.toLowerCase();
  if (v === 'accepted') return 'text-green-400 bg-green-900/30';
  if (v.includes('wrong')) return 'text-red-400 bg-red-900/30';
  if (v.includes('time')) return 'text-yellow-400 bg-yellow-900/30';
  if (v.includes('memory')) return 'text-purple-400 bg-purple-900/30';
  if (v.includes('compile') || v.includes('compilation')) return 'text-orange-400 bg-orange-900/30';
  if (v.includes('runtime')) return 'text-red-400 bg-red-900/30';
  return 'text-gray-400 bg-gray-900/30';
}

/**
 * Get verdict icon
 */
function getVerdictIcon(verdict: string): string {
  const v = verdict.toLowerCase();
  if (v === 'accepted') return '✓';
  if (v.includes('wrong')) return '✗';
  if (v.includes('time')) return '⏱';
  if (v.includes('memory')) return '💾';
  if (v.includes('compile') || v.includes('compilation')) return '🔧';
  if (v.includes('runtime')) return '💥';
  return '?';
}

/**
 * Format date relative to now
 */
function formatRelativeDate(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / (1000 * 60));
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined,
  });
}

/**
 * SubmissionCard - Expandable card showing submission details
 */
export default function SubmissionCard({ submission }: SubmissionCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  const verdictColor = getVerdictColor(submission.verdict);
  const verdictIcon = getVerdictIcon(submission.verdict);
  const languageLabel = LANGUAGE_LABELS[submission.language as Language] || submission.language;

  return (
    <div className="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
      {/* Header - Always visible */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full p-4 flex items-center justify-between hover:bg-gray-750 transition-colors text-left"
      >
        <div className="flex items-center gap-4">
          {/* Verdict badge */}
          <div className={`flex items-center gap-2 px-3 py-1.5 rounded-full ${verdictColor}`}>
            <span className="text-sm">{verdictIcon}</span>
            <span className="text-sm font-medium">{submission.verdict}</span>
          </div>

          {/* Problem title */}
          <div className="flex flex-col">
            <span className="text-white font-medium">{submission.problemTitle}</span>
            <div className="flex items-center gap-2 text-gray-400 text-sm">
              <span>{languageLabel}</span>
              <span>•</span>
              <span>{formatRelativeDate(submission.submittedAt)}</span>
            </div>
          </div>
        </div>

        {/* Metrics */}
        <div className="flex items-center gap-6">
          <div className="text-right">
            <span className="text-gray-400 text-xs block">Time</span>
            <span className="text-white text-sm font-mono">{submission.executionTime}</span>
          </div>
          <div className="text-right">
            <span className="text-gray-400 text-xs block">Memory</span>
            <span className="text-white text-sm font-mono">{submission.memoryUsed}</span>
          </div>
          <div className="text-right">
            <span className="text-gray-400 text-xs block">Tests</span>
            <span className="text-white text-sm font-mono">
              {submission.testsPassed}/{submission.testsTotal}
            </span>
          </div>

          {/* Expand indicator */}
          <span className="text-gray-400 text-lg">
            {isExpanded ? '▼' : '▶'}
          </span>
        </div>
      </button>

      {/* Expanded content */}
      {isExpanded && (
        <div className="border-t border-gray-700 p-4 space-y-4">
          {/* Code preview */}
          {submission.code && (
            <div>
              <h4 className="text-gray-400 text-sm font-medium mb-2">Submitted Code</h4>
              <pre className="bg-gray-900 rounded-lg p-4 text-sm font-mono text-gray-300 overflow-x-auto max-h-64 overflow-y-auto">
                {submission.code}
              </pre>
            </div>
          )}

          {/* Submission details */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="bg-gray-900 rounded-lg p-3">
              <span className="text-gray-400 text-xs block mb-1">Submission ID</span>
              <span className="text-white text-sm font-mono">#{submission.id}</span>
            </div>
            <div className="bg-gray-900 rounded-lg p-3">
              <span className="text-gray-400 text-xs block mb-1">Problem ID</span>
              <span className="text-white text-sm font-mono">#{submission.problemId}</span>
            </div>
            <div className="bg-gray-900 rounded-lg p-3">
              <span className="text-gray-400 text-xs block mb-1">Language</span>
              <span className="text-white text-sm">{languageLabel}</span>
            </div>
            <div className="bg-gray-900 rounded-lg p-3">
              <span className="text-gray-400 text-xs block mb-1">Submitted</span>
              <span className="text-white text-sm">
                {new Date(submission.submittedAt).toLocaleString()}
              </span>
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-3">
            <button className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white text-sm rounded-lg transition-colors">
              View Problem
            </button>
            <button className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded-lg transition-colors">
              Retry Problem
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
