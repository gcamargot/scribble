import { useState } from 'react';
import {
  useLeaderboard,
  METRIC_CONFIG,
  formatMetricValue,
  type MetricType,
  type LeaderboardEntry,
} from '../hooks/useLeaderboard';
import { useAuthStore } from '../stores/authStore';

/**
 * Metric tab button component
 */
function MetricTab({
  metric,
  isActive,
  onClick,
}: {
  metric: MetricType;
  isActive: boolean;
  onClick: () => void;
}) {
  const config = METRIC_CONFIG[metric];

  return (
    <button
      onClick={onClick}
      className={`
        flex items-center gap-2 px-4 py-2.5 rounded-lg font-medium transition-all
        ${isActive
          ? 'bg-blue-600 text-white shadow-lg shadow-blue-600/30'
          : 'bg-gray-800 text-gray-300 hover:bg-gray-700 hover:text-white'
        }
      `}
    >
      <span className="text-lg">{config.icon}</span>
      <span className="hidden sm:inline">{config.label}</span>
    </button>
  );
}

/**
 * Rank badge with different styles based on position
 */
function RankBadge({ rank }: { rank: number }) {
  if (rank === 1) {
    return (
      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-yellow-500 text-black font-bold">
        1
      </div>
    );
  }
  if (rank === 2) {
    return (
      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-gray-300 text-black font-bold">
        2
      </div>
    );
  }
  if (rank === 3) {
    return (
      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-amber-600 text-white font-bold">
        3
      </div>
    );
  }

  return (
    <div className="flex items-center justify-center w-8 h-8 text-gray-400 font-mono">
      {rank}
    </div>
  );
}

/**
 * User avatar with fallback
 */
function UserAvatar({ username, avatarUrl }: { username: string; avatarUrl?: string }) {
  const initials = username
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);

  if (avatarUrl) {
    return (
      <img
        src={avatarUrl}
        alt={username}
        className="w-10 h-10 rounded-full object-cover"
        onError={(e) => {
          // Fallback to initials on error
          e.currentTarget.style.display = 'none';
          e.currentTarget.nextElementSibling?.classList.remove('hidden');
        }}
      />
    );
  }

  return (
    <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-bold text-sm">
      {initials || '?'}
    </div>
  );
}

/**
 * Leaderboard row component
 */
function LeaderboardRow({
  entry,
  metric,
  isCurrentUser,
}: {
  entry: LeaderboardEntry;
  metric: MetricType;
  isCurrentUser: boolean;
}) {
  return (
    <div
      className={`
        flex items-center gap-4 p-4 rounded-lg transition-colors
        ${isCurrentUser
          ? 'bg-blue-900/40 border border-blue-600'
          : 'bg-gray-800 hover:bg-gray-750'
        }
      `}
    >
      {/* Rank */}
      <RankBadge rank={entry.rank} />

      {/* User info */}
      <div className="flex items-center gap-3 flex-1 min-w-0">
        <UserAvatar username={entry.username} avatarUrl={entry.avatar_url} />
        <div className="flex-1 min-w-0">
          <p className={`font-medium truncate ${isCurrentUser ? 'text-blue-300' : 'text-white'}`}>
            {entry.username}
            {isCurrentUser && (
              <span className="ml-2 text-xs bg-blue-600 px-2 py-0.5 rounded-full">You</span>
            )}
          </p>
        </div>
      </div>

      {/* Score */}
      <div className="text-right">
        <span className="text-lg font-bold text-white font-mono">
          {formatMetricValue(metric, entry.metric_value)}
        </span>
      </div>
    </div>
  );
}

/**
 * Pagination component
 */
function Pagination({
  page,
  totalPages,
  onPageChange,
}: {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}) {
  if (totalPages <= 1) return null;

  const pages: (number | 'ellipsis')[] = [];

  // Always show first page
  pages.push(1);

  // Add ellipsis if needed
  if (page > 3) pages.push('ellipsis');

  // Add pages around current
  for (let i = Math.max(2, page - 1); i <= Math.min(totalPages - 1, page + 1); i++) {
    if (!pages.includes(i)) pages.push(i);
  }

  // Add ellipsis if needed
  if (page < totalPages - 2) pages.push('ellipsis');

  // Always show last page
  if (totalPages > 1 && !pages.includes(totalPages)) pages.push(totalPages);

  return (
    <div className="flex items-center justify-center gap-2 mt-6">
      <button
        onClick={() => onPageChange(page - 1)}
        disabled={page === 1}
        className="px-3 py-1.5 rounded bg-gray-700 text-gray-300 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-600 transition-colors"
      >
        Prev
      </button>

      {pages.map((p, i) =>
        p === 'ellipsis' ? (
          <span key={`ellipsis-${i}`} className="text-gray-500 px-2">
            ...
          </span>
        ) : (
          <button
            key={p}
            onClick={() => onPageChange(p)}
            className={`
              w-8 h-8 rounded transition-colors
              ${p === page
                ? 'bg-blue-600 text-white'
                : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
              }
            `}
          >
            {p}
          </button>
        )
      )}

      <button
        onClick={() => onPageChange(page + 1)}
        disabled={page === totalPages}
        className="px-3 py-1.5 rounded bg-gray-700 text-gray-300 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-600 transition-colors"
      >
        Next
      </button>
    </div>
  );
}

/**
 * LeaderboardPage - Rankings with metric tabs and pagination
 */
export default function LeaderboardPage() {
  const [activeMetric, setActiveMetric] = useState<MetricType>('fastest_avg');
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const { entries, total, totalPages, isLoading, isError, error } = useLeaderboard(
    activeMetric,
    page,
    pageSize
  );

  const user = useAuthStore((state) => state.user);

  // Reset page when metric changes
  const handleMetricChange = (metric: MetricType) => {
    setActiveMetric(metric);
    setPage(1);
  };

  const metrics: MetricType[] = [
    'fastest_avg',
    'lowest_memory_avg',
    'problems_solved',
    'longest_streak',
  ];

  return (
    <div className="min-h-screen bg-dark">
      <div className="max-w-4xl mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Leaderboard</h1>
          <p className="text-gray-400">See how you rank against other coders</p>
        </div>

        {/* Metric Tabs */}
        <div className="flex flex-wrap gap-2 mb-6">
          {metrics.map((metric) => (
            <MetricTab
              key={metric}
              metric={metric}
              isActive={activeMetric === metric}
              onClick={() => handleMetricChange(metric)}
            />
          ))}
        </div>

        {/* Active Metric Header */}
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700 mb-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <span className="text-2xl">{METRIC_CONFIG[activeMetric].icon}</span>
              <div>
                <h2 className="text-white font-semibold">{METRIC_CONFIG[activeMetric].label}</h2>
                <p className="text-gray-400 text-sm">
                  {activeMetric === 'fastest_avg' && 'Ranked by average execution time (lower is better)'}
                  {activeMetric === 'lowest_memory_avg' && 'Ranked by average memory usage (lower is better)'}
                  {activeMetric === 'problems_solved' && 'Ranked by total problems solved'}
                  {activeMetric === 'longest_streak' && 'Ranked by longest solving streak'}
                </p>
              </div>
            </div>
            <div className="text-gray-400 text-sm">
              {total.toLocaleString()} ranked users
            </div>
          </div>
        </div>

        {/* Leaderboard List */}
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-2 border-blue-500 border-t-transparent mx-auto mb-4"></div>
              <p className="text-gray-400">Loading leaderboard...</p>
            </div>
          </div>
        ) : isError ? (
          <div className="bg-red-900/50 border border-red-700 rounded-lg p-6 text-center">
            <p className="text-red-300">Failed to load leaderboard</p>
            <p className="text-red-400 text-sm mt-1">{error?.message}</p>
          </div>
        ) : entries.length === 0 ? (
          <div className="bg-gray-800 border border-gray-700 rounded-lg p-12 text-center">
            <div className="text-4xl mb-4">🏆</div>
            <h3 className="text-white font-semibold text-lg mb-2">No rankings yet</h3>
            <p className="text-gray-400">
              Start solving problems to appear on the leaderboard!
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {entries.map((entry) => (
              <LeaderboardRow
                key={entry.id}
                entry={entry}
                metric={activeMetric}
                isCurrentUser={user?.id === String(entry.user_id)}
              />
            ))}
          </div>
        )}

        {/* Pagination */}
        {!isLoading && !isError && entries.length > 0 && (
          <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
        )}
      </div>
    </div>
  );
}
