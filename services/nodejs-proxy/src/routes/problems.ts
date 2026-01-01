import { Router, Request, Response } from 'express';
import NodeCache from 'node-cache';

const router = Router();

// Cache configuration: 1 hour TTL, check every 10 minutes
const problemsCache = new NodeCache({
  stdTTL: 3600,        // 1 hour
  checkperiod: 600,    // Check for expired keys every 10 minutes
  useClones: false     // Don't clone values (better performance)
});

// Go backend URL from environment
const GO_BACKEND_URL = process.env.GO_BACKEND_URL || 'http://localhost:8080';

/**
 * GET /api/problems/daily
 * Returns today's daily challenge with caching
 */
router.get('/daily', async (req: Request, res: Response) => {
  const cacheKey = 'daily_challenge_today';

  // Check cache first
  const cached = problemsCache.get(cacheKey);
  if (cached) {
    return res.json({
      ...cached as object,
      _cached: true
    });
  }

  try {
    // Fetch from Go backend
    const response = await fetch(`${GO_BACKEND_URL}/internal/problems/daily/today`);

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      return res.status(response.status).json(error);
    }

    const data = await response.json();

    // Cache the response for 1 hour
    problemsCache.set(cacheKey, data);

    return res.json({
      ...data,
      _cached: false
    });
  } catch (error) {
    console.error('[Problems] Error fetching daily challenge:', error);
    return res.status(502).json({
      error: 'Failed to fetch daily challenge from backend'
    });
  }
});

/**
 * GET /api/problems/daily/:date
 * Returns daily challenge for a specific date with caching
 */
router.get('/daily/:date', async (req: Request, res: Response) => {
  const { date } = req.params;
  const cacheKey = `daily_challenge_${date}`;

  // Check cache first
  const cached = problemsCache.get(cacheKey);
  if (cached) {
    return res.json({
      ...cached as object,
      _cached: true
    });
  }

  try {
    const response = await fetch(`${GO_BACKEND_URL}/internal/problems/daily/${date}`);

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      return res.status(response.status).json(error);
    }

    const data = await response.json();

    // Cache the response for 1 hour
    problemsCache.set(cacheKey, data);

    return res.json({
      ...data,
      _cached: false
    });
  } catch (error) {
    console.error(`[Problems] Error fetching daily challenge for ${date}:`, error);
    return res.status(502).json({
      error: 'Failed to fetch daily challenge from backend'
    });
  }
});

/**
 * GET /api/problems/:id
 * Returns a specific problem by ID with caching
 */
router.get('/:id', async (req: Request, res: Response) => {
  const { id } = req.params;
  const cacheKey = `problem_${id}`;

  // Check cache first
  const cached = problemsCache.get(cacheKey);
  if (cached) {
    return res.json({
      ...cached as object,
      _cached: true
    });
  }

  try {
    const response = await fetch(`${GO_BACKEND_URL}/internal/problems/${id}`);

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      return res.status(response.status).json(error);
    }

    const data = await response.json();

    // Cache the response for 1 hour
    problemsCache.set(cacheKey, data);

    return res.json({
      ...data,
      _cached: false
    });
  } catch (error) {
    console.error(`[Problems] Error fetching problem ${id}:`, error);
    return res.status(502).json({
      error: 'Failed to fetch problem from backend'
    });
  }
});

/**
 * GET /api/problems/:id/test-cases
 * Returns sample test cases for a problem (no caching - may change)
 */
router.get('/:id/test-cases', async (req: Request, res: Response) => {
  const { id } = req.params;
  // Only return sample test cases unless explicitly requested
  const all = req.query.all === 'true';

  try {
    const url = `${GO_BACKEND_URL}/internal/problems/${id}/test-cases${all ? '?all=true' : ''}`;
    const response = await fetch(url);

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      return res.status(response.status).json(error);
    }

    const data = await response.json();
    return res.json(data);
  } catch (error) {
    console.error(`[Problems] Error fetching test cases for problem ${id}:`, error);
    return res.status(502).json({
      error: 'Failed to fetch test cases from backend'
    });
  }
});

/**
 * Clear cache endpoint (for admin use)
 */
router.post('/cache/clear', (req: Request, res: Response) => {
  // TODO: Add authentication check for admin
  problemsCache.flushAll();
  res.json({ message: 'Cache cleared', keys: problemsCache.keys().length });
});

export default router;
