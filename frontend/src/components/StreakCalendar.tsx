import CalendarHeatmap, { type ReactCalendarHeatmapValue } from 'react-calendar-heatmap';
import 'react-calendar-heatmap/dist/styles.css';
import { useUserActivity } from '../hooks/useSubmissionHistory';

interface StreakCalendarProps {
  currentStreak: number;
  longestStreak: number;
  totalDaysActive: number;
}

// Extended value type for the heatmap
type HeatmapValue = ReactCalendarHeatmapValue<string> & {
  count?: number;
  solved?: boolean;
  attempted?: boolean;
};

/**
 * Get CSS class for a day based on activity
 */
function getClassForValue(value: HeatmapValue | undefined): string {
  if (!value) return 'color-empty';
  if (value.solved) {
    // Intensity based on number of submissions
    if (value.count && value.count >= 5) return 'color-solved-4';
    if (value.count && value.count >= 3) return 'color-solved-3';
    if (value.count && value.count >= 2) return 'color-solved-2';
    return 'color-solved-1';
  }
  if (value.attempted) return 'color-attempted';
  return 'color-empty';
}

/**
 * Format tooltip text for a day
 */
function getTooltip(value: HeatmapValue | undefined): string {
  if (!value || !value.date) return 'No activity';
  const date = new Date(value.date).toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
  });
  if (value.solved) {
    return `${date}: ${value.count ?? 1} submission${(value.count ?? 1) > 1 ? 's' : ''} (solved)`;
  }
  if (value.attempted) {
    return `${date}: ${value.count ?? 1} submission${(value.count ?? 1) > 1 ? 's' : ''} (attempted)`;
  }
  return `${date}: No activity`;
}

/**
 * Month labels for the calendar
 */
const MONTH_LABELS: [string, string, string, string, string, string, string, string, string, string, string, string] =
  ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

/**
 * StreakCalendar - GitHub-style activity heatmap
 */
export default function StreakCalendar({
  currentStreak,
  longestStreak,
  totalDaysActive,
}: StreakCalendarProps) {
  const { activity, isLoading, isError } = useUserActivity();

  // Calculate date range (past year)
  const endDate = new Date();
  const startDate = new Date();
  startDate.setFullYear(startDate.getFullYear() - 1);

  // Transform activity data for the heatmap
  const heatmapValues = activity.map((day) => ({
    date: day.date,
    count: day.count,
    solved: day.solved,
    attempted: day.attempted,
  }));

  if (isLoading) {
    return (
      <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-700 rounded w-40 mb-4" />
          <div className="h-32 bg-gray-700 rounded" />
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
        <p className="text-red-400">Failed to load activity data</p>
      </div>
    );
  }

  return (
    <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
      <h2 className="text-lg font-semibold text-white mb-4">Activity</h2>

      {/* Stats Row */}
      <div className="flex flex-wrap gap-6 mb-6">
        <div className="flex items-center gap-2">
          <span className="text-orange-400 text-lg">🔥</span>
          <div>
            <span className="text-white font-bold">{currentStreak}</span>
            <span className="text-gray-400 text-sm ml-1">current streak</span>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-yellow-400 text-lg">🏆</span>
          <div>
            <span className="text-white font-bold">{longestStreak}</span>
            <span className="text-gray-400 text-sm ml-1">longest streak</span>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-green-400 text-lg">📅</span>
          <div>
            <span className="text-white font-bold">{totalDaysActive}</span>
            <span className="text-gray-400 text-sm ml-1">days active</span>
          </div>
        </div>
      </div>

      {/* Calendar Heatmap */}
      <div className="streak-calendar-wrapper">
        <CalendarHeatmap
          startDate={startDate}
          endDate={endDate}
          values={heatmapValues}
          classForValue={getClassForValue}
          titleForValue={getTooltip}
          showWeekdayLabels
          monthLabels={MONTH_LABELS}
        />
      </div>

      {/* Legend */}
      <div className="flex items-center justify-end gap-2 mt-4 text-xs text-gray-400">
        <span>Less</span>
        <div className="w-3 h-3 rounded-sm bg-gray-700" title="No activity" />
        <div className="w-3 h-3 rounded-sm bg-yellow-600" title="Attempted" />
        <div className="w-3 h-3 rounded-sm bg-green-800" title="1 solved" />
        <div className="w-3 h-3 rounded-sm bg-green-600" title="2 solved" />
        <div className="w-3 h-3 rounded-sm bg-green-500" title="3+ solved" />
        <div className="w-3 h-3 rounded-sm bg-green-400" title="5+ solved" />
        <span>More</span>
      </div>
    </div>
  );
}
