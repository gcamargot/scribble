import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

/**
 * Available leaderboard metric types
 */
export type MetricType =
  | 'fastest_avg'
  | 'lowest_memory_avg'
  | 'problems_solved'
  | 'longest_streak';

/**
 * Metric display configuration
 */
export const METRIC_CONFIG: Record<MetricType, { label: string; unit: string; icon: string }> = {
  fastest_avg: { label: 'Fastest Average', unit: 'ms', icon: '⚡' },
  lowest_memory_avg: { label: 'Lowest Memory', unit: 'MB', icon: '💾' },
  problems_solved: { label: 'Problems Solved', unit: '', icon: '✅' },
  longest_streak: { label: 'Longest Streak', unit: 'days', icon: '🔥' },
};

/**
 * Leaderboard entry with user info
 */
export interface LeaderboardEntry {
  id: number;
  user_id: number;
  metric_type: MetricType;
  metric_value: number;
  rank: number;
  computed_at: string;
  username: string;
  avatar_url?: string;
}

/**
 * Paginated leaderboard response
 */
export interface LeaderboardPage {
  entries: LeaderboardEntry[];
  metric_type: MetricType;
  page: number;
  page_size: number;
  total_pages: number;
  total: number;
}

/**
 * User ranks across all metrics
 */
export interface UserRanks {
  user_id: number;
  ranks: Record<MetricType, { rank: number; value: number } | null>;
}

/**
 * Fetch leaderboard for a specific metric
 */
async function fetchLeaderboard(
  metric: MetricType,
  page: number = 1,
  pageSize: number = 20
): Promise<LeaderboardPage> {
  const response = await axios.get<LeaderboardPage>(
    `/api/leaderboards/${metric}?page=${page}&page_size=${pageSize}`
  );
  return response.data;
}

/**
 * Fetch user's ranks across all metrics
 */
async function fetchUserRanks(userId: number): Promise<UserRanks> {
  const response = await axios.get<UserRanks>(`/api/leaderboards/user/${userId}`);
  return response.data;
}

/**
 * Hook to fetch leaderboard for a specific metric
 *
 * @example
 * const { leaderboard, isLoading } = useLeaderboard('fastest_avg');
 */
export function useLeaderboard(metric: MetricType, page: number = 1, pageSize: number = 20) {
  const {
    data: leaderboard,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: ['leaderboard', metric, page, pageSize],
    queryFn: () => fetchLeaderboard(metric, page, pageSize),
    staleTime: 1000 * 60 * 5, // 5 minutes - leaderboards computed every 5 mins
    gcTime: 1000 * 60 * 30, // 30 minutes
  });

  return {
    leaderboard,
    entries: leaderboard?.entries ?? [],
    total: leaderboard?.total ?? 0,
    totalPages: leaderboard?.total_pages ?? 0,
    isLoading,
    isError,
    error: error as Error | null,
    refetch,
  };
}

/**
 * Hook to fetch user's ranks across all metrics
 *
 * @example
 * const { ranks, isLoading } = useUserRanks(userId);
 */
export function useUserRanks(userId: number | undefined) {
  const {
    data,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: ['userRanks', userId],
    queryFn: () => fetchUserRanks(userId!),
    enabled: !!userId,
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 60,
  });

  return {
    ranks: data?.ranks ?? null,
    isLoading,
    isError,
    error: error as Error | null,
    refetch,
  };
}

/**
 * Format metric value for display
 */
export function formatMetricValue(metric: MetricType, value: number): string {
  const config = METRIC_CONFIG[metric];

  switch (metric) {
    case 'fastest_avg':
      return `${value.toFixed(0)}${config.unit}`;
    case 'lowest_memory_avg':
      return `${value.toFixed(1)}${config.unit}`;
    case 'problems_solved':
      return `${Math.floor(value)}`;
    case 'longest_streak':
      return `${Math.floor(value)} ${config.unit}`;
    default:
      return `${value}`;
  }
}
