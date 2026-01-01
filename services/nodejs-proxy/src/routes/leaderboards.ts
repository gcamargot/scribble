import { Router, Request, Response } from 'express';
import NodeCache from 'node-cache';

const router = Router();

// Cache configuration: 1 minute TTL for leaderboard data
// Short TTL since leaderboards are computed every 5 minutes
const leaderboardCache = new NodeCache({
  stdTTL: 60,         // 1 minute
  checkperiod: 30,    // Check for expired keys every 30 seconds
  useClones: false    // Don't clone values (better performance)
});

// Go backend URL from environment
const GO_BACKEND_URL = process.env.GO_BACKEND_URL || 'http://localhost:8080';

// Valid metric types
const VALID_METRICS = ['fastest_avg', 'lowest_memory_avg', 'problems_solved', 'longest_streak'];

/**
 * GET /api/leaderboards/metrics
 * Returns list of available metric types
 */
router.get('/metrics', async (_req: Request, res: Response) => {
  res.json({
    metrics: VALID_METRICS,
    descriptions: {
      fastest_avg: 'Average execution time (lower is better)',
      lowest_memory_avg: 'Average memory usage (lower is better)',
      problems_solved: 'Total problems solved (higher is better)',
      longest_streak: 'Longest daily challenge streak (higher is better)'
    }
  });
});

/**
 * GET /api/leaderboards/:metric
 * Returns paginated leaderboard for a specific metric
 * Query params: page (default 1), page_size (default 20, max 100)
 */
router.get('/:metric', async (req: Request, res: Response) => {
  const { metric } = req.params;

  // Validate metric type
  if (!VALID_METRICS.includes(metric)) {
    return res.status(400).json({
      error: 'Invalid metric type',
      valid_metrics: VALID_METRICS
    });
  }

  // Parse pagination params
  const page = Math.max(1, parseInt(req.query.page as string) || 1);
  const pageSize = Math.min(100, Math.max(1, parseInt(req.query.page_size as string) || 20));

  const cacheKey = `leaderboard_${metric}_p${page}_s${pageSize}`;

  // Check cache first
  const cached = leaderboardCache.get(cacheKey);
  if (cached) {
    return res.json({
      ...cached as object,
      _cached: true
    });
  }

  try {
    const url = `${GO_BACKEND_URL}/internal/leaderboards/${metric}?page=${page}&page_size=${pageSize}`;
    const response = await fetch(url, {
      headers: {
        'X-Internal-Auth': process.env.INTERNAL_AUTH_SECRET || ''
      }
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      return res.status(response.status).json(error);
    }

    const data = await response.json();

    // Cache the response for 1 minute
    leaderboardCache.set(cacheKey, data);

    return res.json({
      ...data,
      _cached: false
    });
  } catch (error) {
    console.error(`[Leaderboards] Error fetching ${metric} leaderboard:`, error);
    return res.status(502).json({
      error: 'Failed to fetch leaderboard from backend'
    });
  }
});

// Extend Request type with user context
interface AuthenticatedRequest extends Request {
  user?: {
    userId: string;
    username: string;
  };
}

/**
 * GET /api/leaderboards/me
 * Returns current user's ranks across all metrics
 * Requires authentication
 */
router.get('/me', async (req: Request, res: Response) => {
  const authReq = req as AuthenticatedRequest;

  if (!authReq.user?.userId) {
    return res.status(401).json({
      error: 'Authentication required'
    });
  }

  const userId = authReq.user.userId;
  const cacheKey = `user_ranks_${userId}`;

  // Check cache first
  const cached = leaderboardCache.get(cacheKey);
  if (cached) {
    return res.json({
      ...cached as object,
      _cached: true
    });
  }

  try {
    const url = `${GO_BACKEND_URL}/internal/leaderboards/user/${userId}`;
    const response = await fetch(url, {
      headers: {
        'X-Internal-Auth': process.env.INTERNAL_AUTH_SECRET || '',
        'X-User-Id': userId,
        'X-Username': authReq.user.username || ''
      }
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      return res.status(response.status).json(error);
    }

    const data = await response.json();

    // Cache user ranks for 1 minute
    leaderboardCache.set(cacheKey, data);

    return res.json({
      ...data,
      _cached: false
    });
  } catch (error) {
    console.error(`[Leaderboards] Error fetching user ${userId} ranks:`, error);
    return res.status(502).json({
      error: 'Failed to fetch user ranks from backend'
    });
  }
});

/**
 * Clear cache endpoint (for admin use)
 */
router.post('/cache/clear', (req: Request, res: Response) => {
  // TODO: Add authentication check for admin
  leaderboardCache.flushAll();
  res.json({ message: 'Leaderboard cache cleared', keys: leaderboardCache.keys().length });
});

export default router;
