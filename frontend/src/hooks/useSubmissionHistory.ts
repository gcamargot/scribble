import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import axios from 'axios';
import type { SubmissionResult } from '../components/ResultPanel';

/**
 * Submission history entry
 */
export interface SubmissionHistoryEntry {
  id: number;
  problemId: number;
  problemTitle: string;
  language: string;
  status: string;
  verdict: string;
  executionTime: string;
  memoryUsed: string;
  testsPassed: number;
  testsTotal: number;
  submittedAt: string;
  code?: string;
  result?: SubmissionResult;
}

/**
 * Paginated response from history endpoint
 */
interface SubmissionHistoryResponse {
  submissions: SubmissionHistoryEntry[];
  page: number;
  pageSize: number;
  totalCount: number;
  totalPages: number;
  hasMore: boolean;
}

/**
 * Filter options for submission history
 */
export interface SubmissionHistoryFilters {
  problemId?: number;
  status?: 'accepted' | 'wrong_answer' | 'time_limit' | 'runtime_error' | 'compile_error';
  language?: string;
}

/**
 * Fetch submission history with pagination
 */
async function fetchSubmissionHistory(
  page: number,
  pageSize: number = 20,
  filters?: SubmissionHistoryFilters
): Promise<SubmissionHistoryResponse> {
  const params = new URLSearchParams();
  params.set('page', String(page));
  params.set('page_size', String(pageSize));

  if (filters?.problemId) params.set('problem_id', String(filters.problemId));
  if (filters?.status) params.set('status', filters.status);
  if (filters?.language) params.set('language', filters.language);

  const response = await axios.get<SubmissionHistoryResponse>(
    `/api/submissions/history?${params.toString()}`
  );
  return response.data;
}

/**
 * Submission statistics
 */
export interface SubmissionStats {
  totalSubmissions: number;
  acceptedSubmissions: number;
  acceptanceRate: number;
  totalProblems: number;
  solvedProblems: number;
  currentStreak: number;
  longestStreak: number;
  lastSolvedDate: string | null;
  languageBreakdown: Record<string, number>;
  difficultyBreakdown: {
    easy: number;
    medium: number;
    hard: number;
  };
  averageMetrics: {
    executionTime: number;
    memoryUsed: number;
  };
}

/**
 * Hook to fetch submission history with infinite scrolling
 *
 * @example
 * const { submissions, fetchNextPage, hasNextPage, isLoading } = useSubmissionHistory();
 */
export function useSubmissionHistory(filters?: SubmissionHistoryFilters) {
  const {
    data,
    isLoading,
    isError,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    refetch,
  } = useInfiniteQuery({
    queryKey: ['submissionHistory', filters],
    queryFn: ({ pageParam = 1 }) => fetchSubmissionHistory(pageParam, 20, filters),
    getNextPageParam: (lastPage) => {
      if (!lastPage.hasMore) return undefined;
      return lastPage.page + 1;
    },
    initialPageParam: 1,
    staleTime: 1000 * 60 * 5, // 5 minutes
    gcTime: 1000 * 60 * 30, // 30 minutes
  });

  // Flatten pages into single array
  const submissions = data?.pages.flatMap(page => page.submissions) ?? [];
  const totalCount = data?.pages[0]?.totalCount ?? 0;

  return {
    submissions,
    totalCount,
    isLoading,
    isError,
    error: error as Error | null,
    fetchNextPage,
    hasNextPage: !!hasNextPage,
    isFetchingNextPage,
    refetch,
  };
}

/**
 * Hook to fetch submission statistics
 *
 * @example
 * const { stats, isLoading } = useSubmissionStats();
 */
export function useSubmissionStats() {
  const {
    data: stats,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: ['submissionStats'],
    queryFn: async () => {
      const response = await axios.get<SubmissionStats>('/api/submissions/stats');
      return response.data;
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
    gcTime: 1000 * 60 * 60, // 1 hour
  });

  return {
    stats,
    isLoading,
    isError,
    error: error as Error | null,
    refetch,
  };
}

/**
 * Daily activity entry for calendar heatmap
 */
export interface DailyActivity {
  date: string; // YYYY-MM-DD
  count: number; // Number of submissions
  solved: boolean; // Had at least one accepted submission
  attempted: boolean; // Had submissions but none accepted
}

/**
 * Fetch user activity for the past year
 */
async function fetchUserActivity(): Promise<DailyActivity[]> {
  const response = await axios.get<{ activity: DailyActivity[] }>('/api/submissions/activity');
  return response.data.activity;
}

/**
 * Hook to fetch user daily activity for calendar heatmap
 */
export function useUserActivity() {
  const {
    data: activity,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: ['userActivity'],
    queryFn: fetchUserActivity,
    staleTime: 1000 * 60 * 60, // 1 hour
    gcTime: 1000 * 60 * 60 * 24, // 24 hours
  });

  return {
    activity: activity ?? [],
    isLoading,
    isError,
    error: error as Error | null,
    refetch,
  };
}

/**
 * Hook to fetch a single submission by ID
 */
export function useSubmission(submissionId: number | undefined) {
  const {
    data: submission,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: ['submission', submissionId],
    queryFn: async () => {
      const response = await axios.get<SubmissionHistoryEntry>(
        `/api/submissions/${submissionId}`
      );
      return response.data;
    },
    enabled: !!submissionId,
    staleTime: 1000 * 60 * 60, // 1 hour (submissions don't change)
    gcTime: 1000 * 60 * 60 * 24, // 24 hours
  });

  return {
    submission,
    isLoading,
    isError,
    error: error as Error | null,
    refetch,
  };
}
