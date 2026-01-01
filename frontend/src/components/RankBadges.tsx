import { useUserRanks, METRIC_CONFIG, formatMetricValue, type MetricType } from '../hooks/useLeaderboard';

interface RankBadgesProps {
  userId: number;
}

/**
 * Individual rank badge component
 */
function RankBadge({
  metric,
  rank,
  value,
}: {
  metric: MetricType;
  rank: number;
  value: number;
}) {
  const config = METRIC_CONFIG[metric];

  // Determine badge color based on rank
  const getBadgeColor = (rank: number): string => {
    if (rank === 1) return 'from-yellow-500 to-amber-600';
    if (rank === 2) return 'from-gray-300 to-gray-400';
    if (rank === 3) return 'from-amber-600 to-amber-700';
    if (rank <= 10) return 'from-blue-500 to-blue-600';
    if (rank <= 50) return 'from-purple-500 to-purple-600';
    return 'from-gray-600 to-gray-700';
  };

  const getRankLabel = (rank: number): string => {
    if (rank === 1) return '1st';
    if (rank === 2) return '2nd';
    if (rank === 3) return '3rd';
    return `#${rank}`;
  };

  return (
    <a
      href={`/leaderboard?metric=${metric}`}
      className="group block"
    >
      <div className="bg-gray-800 rounded-lg p-4 border border-gray-700 hover:border-gray-600 transition-all hover:shadow-lg">
        <div className="flex items-center gap-3 mb-3">
          <span className="text-2xl">{config.icon}</span>
          <div className="flex-1">
            <p className="text-white font-medium text-sm">{config.label}</p>
            <p className="text-gray-400 text-xs">
              {formatMetricValue(metric, value)}
            </p>
          </div>
        </div>

        {/* Rank badge */}
        <div
          className={`
            inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full
            bg-gradient-to-r ${getBadgeColor(rank)}
            text-white font-bold text-sm shadow-md
            group-hover:scale-105 transition-transform
          `}
        >
          <span>{getRankLabel(rank)}</span>
          {rank <= 3 && (
            <span className="text-base">
              {rank === 1 ? '🥇' : rank === 2 ? '🥈' : '🥉'}
            </span>
          )}
        </div>

        {/* View on leaderboard link */}
        <p className="text-gray-500 text-xs mt-2 group-hover:text-blue-400 transition-colors">
          View on leaderboard →
        </p>
      </div>
    </a>
  );
}

/**
 * Loading skeleton for rank badge
 */
function RankBadgeSkeleton() {
  return (
    <div className="bg-gray-800 rounded-lg p-4 border border-gray-700 animate-pulse">
      <div className="flex items-center gap-3 mb-3">
        <div className="w-8 h-8 bg-gray-700 rounded" />
        <div className="flex-1">
          <div className="h-4 bg-gray-700 rounded w-24 mb-1" />
          <div className="h-3 bg-gray-700 rounded w-16" />
        </div>
      </div>
      <div className="h-8 bg-gray-700 rounded-full w-20" />
    </div>
  );
}

/**
 * RankBadges - Display user's rank for each metric
 */
export default function RankBadges({ userId }: RankBadgesProps) {
  const { ranks, isLoading, isError } = useUserRanks(userId);

  const metrics: MetricType[] = [
    'fastest_avg',
    'lowest_memory_avg',
    'problems_solved',
    'longest_streak',
  ];

  if (isLoading) {
    return (
      <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
        <h2 className="text-lg font-semibold text-white mb-4">Your Rankings</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {metrics.map((metric) => (
            <RankBadgeSkeleton key={metric} />
          ))}
        </div>
      </div>
    );
  }

  if (isError || !ranks) {
    return (
      <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
        <h2 className="text-lg font-semibold text-white mb-4">Your Rankings</h2>
        <p className="text-gray-400">
          Solve some problems to appear on the leaderboards!
        </p>
      </div>
    );
  }

  // Check if user has any ranks
  const hasRanks = metrics.some((m) => ranks[m] !== null);

  if (!hasRanks) {
    return (
      <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
        <h2 className="text-lg font-semibold text-white mb-4">Your Rankings</h2>
        <div className="text-center py-8">
          <div className="text-4xl mb-2">🏆</div>
          <p className="text-gray-400">No rankings yet</p>
          <p className="text-gray-500 text-sm">
            Solve problems to appear on the leaderboards!
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
      <h2 className="text-lg font-semibold text-white mb-4">Your Rankings</h2>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {metrics.map((metric) => {
          const rankData = ranks[metric];
          if (!rankData) return null;

          return (
            <RankBadge
              key={metric}
              metric={metric}
              rank={rankData.rank}
              value={rankData.value}
            />
          );
        })}
      </div>
    </div>
  );
}
