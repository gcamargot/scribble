import { useState, useCallback } from 'react';
import { useSubmissionHistory, useSubmissionStats, type SubmissionHistoryFilters } from '../hooks/useSubmissionHistory';
import SubmissionCard from '../components/SubmissionCard';
import { LANGUAGE_LABELS } from '../constants/starterCode';

/**
 * Status filter options
 */
const STATUS_OPTIONS = [
  { value: '', label: 'All Statuses' },
  { value: 'accepted', label: 'Accepted' },
  { value: 'wrong_answer', label: 'Wrong Answer' },
  { value: 'time_limit', label: 'Time Limit Exceeded' },
  { value: 'runtime_error', label: 'Runtime Error' },
  { value: 'compile_error', label: 'Compilation Error' },
] as const;

/**
 * Language filter options
 */
const LANGUAGE_OPTIONS = [
  { value: '', label: 'All Languages' },
  ...Object.entries(LANGUAGE_LABELS).map(([value, label]) => ({ value, label })),
] as const;

/**
 * Stats card component
 */
function StatsCard({ label, value, subtext }: { label: string; value: string | number; subtext?: string }) {
  return (
    <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
      <span className="text-gray-400 text-sm block mb-1">{label}</span>
      <span className="text-white text-2xl font-bold">{value}</span>
      {subtext && <span className="text-gray-500 text-xs block mt-1">{subtext}</span>}
    </div>
  );
}

/**
 * HistoryPage - Submission history with filters and pagination
 */
export default function HistoryPage() {
  const [filters, setFilters] = useState<SubmissionHistoryFilters>({});

  const {
    submissions,
    totalCount,
    isLoading,
    isError,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useSubmissionHistory(filters);

  const { stats, isLoading: isStatsLoading } = useSubmissionStats();

  // Handle filter changes
  const handleStatusChange = useCallback((status: string) => {
    setFilters(prev => ({
      ...prev,
      status: status ? status as SubmissionHistoryFilters['status'] : undefined,
    }));
  }, []);

  const handleLanguageChange = useCallback((language: string) => {
    setFilters(prev => ({
      ...prev,
      language: language || undefined,
    }));
  }, []);

  const clearFilters = useCallback(() => {
    setFilters({});
  }, []);

  const hasActiveFilters = filters.status || filters.language || filters.problemId;

  return (
    <div className="min-h-screen bg-dark">
      <div className="max-w-6xl mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Submission History</h1>
          <p className="text-gray-400">View and analyze your past submissions</p>
        </div>

        {/* Stats Section */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
          <StatsCard
            label="Total Submissions"
            value={isStatsLoading ? '-' : stats?.totalSubmissions ?? 0}
          />
          <StatsCard
            label="Acceptance Rate"
            value={isStatsLoading ? '-' : `${stats?.acceptanceRate ?? 0}%`}
            subtext={`${stats?.acceptedSubmissions ?? 0} accepted`}
          />
          <StatsCard
            label="Problems Solved"
            value={isStatsLoading ? '-' : `${stats?.solvedProblems ?? 0}/${stats?.totalProblems ?? 0}`}
          />
          <StatsCard
            label="Current Streak"
            value={isStatsLoading ? '-' : `${stats?.currentStreak ?? 0} days`}
            subtext={`Best: ${stats?.longestStreak ?? 0} days`}
          />
        </div>

        {/* Filters */}
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700 mb-6">
          <div className="flex flex-wrap items-center gap-4">
            <div className="flex items-center gap-2">
              <label className="text-gray-400 text-sm">Status:</label>
              <select
                value={filters.status || ''}
                onChange={(e) => handleStatusChange(e.target.value)}
                className="bg-gray-700 text-white rounded px-3 py-1.5 text-sm border border-gray-600 focus:outline-none focus:border-blue-500"
              >
                {STATUS_OPTIONS.map(option => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>

            <div className="flex items-center gap-2">
              <label className="text-gray-400 text-sm">Language:</label>
              <select
                value={filters.language || ''}
                onChange={(e) => handleLanguageChange(e.target.value)}
                className="bg-gray-700 text-white rounded px-3 py-1.5 text-sm border border-gray-600 focus:outline-none focus:border-blue-500"
              >
                {LANGUAGE_OPTIONS.map(option => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>

            {hasActiveFilters && (
              <button
                onClick={clearFilters}
                className="text-blue-400 hover:text-blue-300 text-sm transition-colors"
              >
                Clear filters
              </button>
            )}

            <div className="ml-auto text-gray-400 text-sm">
              {totalCount} submission{totalCount !== 1 ? 's' : ''}
            </div>
          </div>
        </div>

        {/* Submissions List */}
        <div className="space-y-4">
          {isLoading && submissions.length === 0 ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-center">
                <div className="animate-spin rounded-full h-8 w-8 border-2 border-blue-500 border-t-transparent mx-auto mb-4"></div>
                <p className="text-gray-400">Loading submissions...</p>
              </div>
            </div>
          ) : isError ? (
            <div className="bg-red-900/50 border border-red-700 rounded-lg p-6 text-center">
              <p className="text-red-300">Failed to load submissions</p>
              <p className="text-red-400 text-sm mt-1">{error?.message}</p>
            </div>
          ) : submissions.length === 0 ? (
            <div className="bg-gray-800 border border-gray-700 rounded-lg p-12 text-center">
              <div className="text-4xl mb-4">📝</div>
              <h3 className="text-white font-semibold text-lg mb-2">No submissions yet</h3>
              <p className="text-gray-400 mb-4">
                {hasActiveFilters
                  ? 'No submissions match your filters. Try adjusting them.'
                  : "Start solving problems to see your submission history here!"}
              </p>
              {hasActiveFilters && (
                <button
                  onClick={clearFilters}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
                >
                  Clear filters
                </button>
              )}
            </div>
          ) : (
            <>
              {submissions.map((submission) => (
                <SubmissionCard key={submission.id} submission={submission} />
              ))}

              {/* Load More Button */}
              {hasNextPage && (
                <div className="text-center pt-4">
                  <button
                    onClick={() => fetchNextPage()}
                    disabled={isFetchingNextPage}
                    className="px-6 py-2.5 bg-gray-700 hover:bg-gray-600 disabled:bg-gray-800 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                  >
                    {isFetchingNextPage ? (
                      <span className="flex items-center gap-2">
                        <div className="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent"></div>
                        Loading...
                      </span>
                    ) : (
                      'Load More'
                    )}
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
