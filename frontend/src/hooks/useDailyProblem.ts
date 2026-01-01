import { useQuery } from '@tanstack/react-query';
import axios from 'axios';
import { TWO_SUM_PROBLEM, type Problem } from '../constants/problems';

/**
 * API response type for daily problem endpoint
 */
interface DailyProblemResponse extends Problem {
  date?: string;
  _cached?: boolean;
}

/**
 * Fetch the daily problem from the API
 */
async function fetchDailyProblem(): Promise<Problem> {
  const response = await axios.get<DailyProblemResponse>('/api/problems/daily');
  return response.data;
}

/**
 * Hook to fetch the daily coding problem
 *
 * Uses React Query for caching and automatic refetching.
 * Falls back to hardcoded problem if API is unavailable.
 *
 * @example
 * const { problem, isLoading, error } = useDailyProblem();
 */
export function useDailyProblem() {
  const {
    data: problem,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: ['dailyProblem'],
    queryFn: fetchDailyProblem,
    staleTime: 1000 * 60 * 60, // 1 hour - matches backend cache
    gcTime: 1000 * 60 * 60 * 24, // 24 hours
    retry: 2,
    // Use fallback data while loading or on error
    placeholderData: TWO_SUM_PROBLEM,
  });

  return {
    problem: problem ?? TWO_SUM_PROBLEM,
    isLoading,
    isError,
    error: error as Error | null,
    refetch,
    // Helper to check if using fallback data
    isFallback: !problem || problem.id === TWO_SUM_PROBLEM.id,
  };
}

/**
 * Hook to fetch a specific problem by date
 *
 * @param date - Date string in YYYY-MM-DD format
 */
export function useProblemByDate(date: string) {
  const {
    data: problem,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: ['problem', 'daily', date],
    queryFn: async () => {
      const response = await axios.get<DailyProblemResponse>(
        `/api/problems/daily/${date}`
      );
      return response.data;
    },
    staleTime: 1000 * 60 * 60, // 1 hour
    gcTime: 1000 * 60 * 60 * 24, // 24 hours
    retry: 2,
    enabled: !!date, // Only fetch if date is provided
  });

  return {
    problem,
    isLoading,
    isError,
    error: error as Error | null,
    refetch,
  };
}

/**
 * Hook to fetch a specific problem by ID
 *
 * @param problemId - The problem ID
 */
export function useProblemById(problemId: number | undefined) {
  const {
    data: problem,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: ['problem', problemId],
    queryFn: async () => {
      const response = await axios.get<Problem>(`/api/problems/${problemId}`);
      return response.data;
    },
    staleTime: 1000 * 60 * 60, // 1 hour
    gcTime: 1000 * 60 * 60 * 24, // 24 hours
    retry: 2,
    enabled: !!problemId, // Only fetch if problemId is provided
  });

  return {
    problem,
    isLoading,
    isError,
    error: error as Error | null,
    refetch,
  };
}
