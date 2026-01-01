import { useState } from 'react';
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Tooltip,
  Legend,
} from 'recharts';
import { useSubmissionStats, useSubmissionHistory } from '../hooks/useSubmissionHistory';
import { useAuthStore } from '../stores/authStore';
import StreakBadge from '../components/StreakBadge';
import SubmissionCard from '../components/SubmissionCard';
import { LANGUAGE_LABELS, type Language } from '../constants/starterCode';

/**
 * Color palette for charts
 */
const LANGUAGE_COLORS: Record<string, string> = {
  python: '#3776AB',
  javascript: '#F7DF1E',
  typescript: '#3178C6',
  java: '#ED8B00',
  cpp: '#00599C',
  go: '#00ADD8',
  rust: '#DEA584',
};

const DIFFICULTY_COLORS = {
  easy: '#22c55e',
  medium: '#eab308',
  hard: '#ef4444',
};

/**
 * Stats card component
 */
function StatCard({
  label,
  value,
  subtext,
  icon,
}: {
  label: string;
  value: string | number;
  subtext?: string;
  icon?: string;
}) {
  return (
    <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
      <div className="flex items-center gap-2 mb-2">
        {icon && <span className="text-xl">{icon}</span>}
        <span className="text-gray-400 text-sm">{label}</span>
      </div>
      <span className="text-white text-2xl font-bold block">{value}</span>
      {subtext && <span className="text-gray-500 text-xs mt-1 block">{subtext}</span>}
    </div>
  );
}

/**
 * Difficulty progress bar
 */
function DifficultyProgress({
  difficulty,
  solved,
  total,
  color,
}: {
  difficulty: string;
  solved: number;
  total: number;
  color: string;
}) {
  const percentage = total > 0 ? (solved / total) * 100 : 0;

  return (
    <div className="flex items-center gap-4">
      <div className="w-20 text-sm font-medium capitalize" style={{ color }}>
        {difficulty}
      </div>
      <div className="flex-1">
        <div className="h-3 bg-gray-700 rounded-full overflow-hidden">
          <div
            className="h-full rounded-full transition-all duration-500"
            style={{ width: `${percentage}%`, backgroundColor: color }}
          />
        </div>
      </div>
      <div className="w-16 text-right text-gray-300 text-sm font-mono">
        {solved}/{total}
      </div>
    </div>
  );
}

/**
 * Language pie chart component
 */
function LanguageChart({
  breakdown,
}: {
  breakdown: Record<string, number>;
}) {
  const data = Object.entries(breakdown)
    .filter(([, count]) => count > 0)
    .map(([lang, count]) => ({
      name: LANGUAGE_LABELS[lang as Language] || lang,
      value: count,
      color: LANGUAGE_COLORS[lang] || '#6b7280',
    }));

  if (data.length === 0) {
    return (
      <div className="flex items-center justify-center h-48 text-gray-500">
        No submissions yet
      </div>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={200}>
      <PieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          innerRadius={40}
          outerRadius={70}
          paddingAngle={2}
          dataKey="value"
        >
          {data.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={entry.color} />
          ))}
        </Pie>
        <Tooltip
          contentStyle={{
            backgroundColor: '#1f2937',
            border: '1px solid #374151',
            borderRadius: '8px',
          }}
          itemStyle={{ color: '#fff' }}
        />
        <Legend
          wrapperStyle={{ fontSize: '12px' }}
          formatter={(value) => <span style={{ color: '#9ca3af' }}>{value}</span>}
        />
      </PieChart>
    </ResponsiveContainer>
  );
}

/**
 * User avatar component
 */
function ProfileAvatar({
  username,
  avatarUrl,
}: {
  username: string;
  avatarUrl?: string | null;
}) {
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
        className="w-24 h-24 rounded-full object-cover border-4 border-gray-700"
      />
    );
  }

  return (
    <div className="w-24 h-24 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-bold text-3xl border-4 border-gray-700">
      {initials || '?'}
    </div>
  );
}

/**
 * ProfilePage - User profile with stats and charts
 */
export default function ProfilePage() {
  const user = useAuthStore((state) => state.user);
  const { stats, isLoading: isStatsLoading } = useSubmissionStats();
  const { submissions, isLoading: isHistoryLoading } = useSubmissionHistory();
  const [showAllRecent, setShowAllRecent] = useState(false);

  // Get 5 most recent submissions (or all if expanded)
  const recentSubmissions = showAllRecent ? submissions.slice(0, 10) : submissions.slice(0, 5);

  // Build Discord avatar URL
  const avatarUrl = user?.avatar
    ? `https://cdn.discordapp.com/avatars/${user.id}/${user.avatar}.png?size=128`
    : null;

  return (
    <div className="min-h-screen bg-dark">
      <div className="max-w-6xl mx-auto px-4 py-8">
        {/* Profile Header */}
        <div className="bg-gray-800 rounded-lg p-6 border border-gray-700 mb-8">
          <div className="flex flex-col md:flex-row items-center gap-6">
            <ProfileAvatar
              username={user?.username || 'User'}
              avatarUrl={avatarUrl}
            />
            <div className="flex-1 text-center md:text-left">
              <h1 className="text-3xl font-bold text-white mb-2">
                {user?.username || 'User'}
              </h1>
              <p className="text-gray-400">
                Solving problems one day at a time
              </p>
            </div>
            {/* Streak Badge */}
            <div className="flex flex-col items-center">
              {isStatsLoading ? (
                <div className="animate-pulse bg-gray-700 rounded-full h-12 w-32" />
              ) : (
                <StreakBadge
                  currentStreak={stats?.currentStreak ?? 0}
                  longestStreak={stats?.longestStreak ?? 0}
                  lastSolvedDate={stats?.lastSolvedDate}
                  size="lg"
                />
              )}
            </div>
          </div>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
          <StatCard
            label="Total Submissions"
            value={isStatsLoading ? '-' : stats?.totalSubmissions ?? 0}
            icon="📝"
          />
          <StatCard
            label="Acceptance Rate"
            value={isStatsLoading ? '-' : `${stats?.acceptanceRate ?? 0}%`}
            subtext={`${stats?.acceptedSubmissions ?? 0} accepted`}
            icon="✅"
          />
          <StatCard
            label="Avg. Execution Time"
            value={isStatsLoading ? '-' : `${stats?.averageMetrics?.executionTime?.toFixed(0) ?? 0}ms`}
            icon="⚡"
          />
          <StatCard
            label="Avg. Memory"
            value={isStatsLoading ? '-' : `${stats?.averageMetrics?.memoryUsed?.toFixed(1) ?? 0}MB`}
            icon="💾"
          />
        </div>

        {/* Two Column Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Left Column */}
          <div className="space-y-8">
            {/* Problems by Difficulty */}
            <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
              <h2 className="text-lg font-semibold text-white mb-4">Problems Solved</h2>
              <div className="space-y-4">
                <DifficultyProgress
                  difficulty="Easy"
                  solved={stats?.difficultyBreakdown?.easy ?? 0}
                  total={30} // TODO: Get from API
                  color={DIFFICULTY_COLORS.easy}
                />
                <DifficultyProgress
                  difficulty="Medium"
                  solved={stats?.difficultyBreakdown?.medium ?? 0}
                  total={50}
                  color={DIFFICULTY_COLORS.medium}
                />
                <DifficultyProgress
                  difficulty="Hard"
                  solved={stats?.difficultyBreakdown?.hard ?? 0}
                  total={20}
                  color={DIFFICULTY_COLORS.hard}
                />
              </div>
              <div className="mt-4 pt-4 border-t border-gray-700">
                <div className="flex items-center justify-between">
                  <span className="text-gray-400">Total Solved</span>
                  <span className="text-white font-bold">
                    {stats?.solvedProblems ?? 0}/{stats?.totalProblems ?? 100}
                  </span>
                </div>
              </div>
            </div>

            {/* Language Breakdown */}
            <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
              <h2 className="text-lg font-semibold text-white mb-4">Languages Used</h2>
              <LanguageChart breakdown={stats?.languageBreakdown ?? {}} />
            </div>
          </div>

          {/* Right Column */}
          <div className="space-y-8">
            {/* Recent Submissions */}
            <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold text-white">Recent Submissions</h2>
                {submissions.length > 5 && (
                  <button
                    onClick={() => setShowAllRecent(!showAllRecent)}
                    className="text-blue-400 hover:text-blue-300 text-sm transition-colors"
                  >
                    {showAllRecent ? 'Show less' : 'Show more'}
                  </button>
                )}
              </div>

              {isHistoryLoading ? (
                <div className="space-y-3">
                  {[...Array(3)].map((_, i) => (
                    <div key={i} className="animate-pulse bg-gray-700 rounded-lg h-16" />
                  ))}
                </div>
              ) : recentSubmissions.length === 0 ? (
                <div className="text-center py-8">
                  <div className="text-4xl mb-2">📝</div>
                  <p className="text-gray-400">No submissions yet</p>
                  <p className="text-gray-500 text-sm">Start solving problems!</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {recentSubmissions.map((submission) => (
                    <SubmissionCard key={submission.id} submission={submission} />
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
