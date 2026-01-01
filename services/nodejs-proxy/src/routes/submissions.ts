import { Router, Request, Response, NextFunction, RequestHandler } from 'express';
import { submissionsProxy } from '../middleware/proxy.js';

const router = Router();

// Check if we should use the real backend or mock
const USE_MOCK_BACKEND = process.env.USE_MOCK_BACKEND === 'true';

interface SubmissionRequest {
  code: string;
  language: string;
  problemId: number;
}

/**
 * POST /api/submissions
 *
 * Submit code for execution.
 * In production: proxies to Go backend /internal/execute
 * In development with USE_MOCK_BACKEND=true: returns mock response
 */
const handleSubmission: RequestHandler = (req: Request, res: Response, next: NextFunction): void => {
  const { code, language, problemId } = req.body as SubmissionRequest;

  // Basic validation
  if (!code || !language || !problemId) {
    res.status(400).json({
      error: 'Missing required fields: code, language, problemId'
    });
    return;
  }

  // If mock mode enabled, return fake response
  if (USE_MOCK_BACKEND) {
    setTimeout(() => {
      res.json({
        success: true,
        status: 'Accepted',
        verdict: 'Accepted',
        executionTime: '45ms',
        memoryUsed: '14.2MB',
        testsPassed: 3,
        testsTotal: 3,
        percentile: {
          time: 78,
          memory: 65
        },
        message: 'All test cases passed!',
        submission: {
          id: Math.floor(Math.random() * 100000),
          problemId,
          language,
          timestamp: new Date().toISOString(),
          codeLength: code.length
        }
      });
    }, 800);
    return;
  }

  // Forward to Go backend via proxy
  next();
};

// Apply validation then proxy
router.post('/', handleSubmission, submissionsProxy);

// Go backend URL from environment
const GO_BACKEND_URL = process.env.GO_BACKEND_URL || 'http://localhost:8080';

// Extend Request type with user context (set by auth middleware)
interface AuthenticatedRequest extends Request {
  user?: {
    userId: string;
    username: string;
  };
}

/**
 * GET /api/submissions/history
 *
 * Get current user's submission history (paginated)
 * Query params: page, page_size, problem_id, status, language
 * Requires authentication
 */
router.get('/history', async (req: Request, res: Response) => {
  const authReq = req as AuthenticatedRequest;

  // Check if user is authenticated
  if (!authReq.user?.userId) {
    return res.status(401).json({
      error: 'Authentication required'
    });
  }

  const userId = authReq.user.userId;

  // Build query string from request params
  const queryParams = new URLSearchParams();
  if (req.query.page) queryParams.set('page', req.query.page as string);
  if (req.query.page_size) queryParams.set('page_size', req.query.page_size as string);
  if (req.query.problem_id) queryParams.set('problem_id', req.query.problem_id as string);
  if (req.query.status) queryParams.set('status', req.query.status as string);
  if (req.query.language) queryParams.set('language', req.query.language as string);

  const queryString = queryParams.toString();
  const url = `${GO_BACKEND_URL}/internal/submissions/user/${userId}${queryString ? '?' + queryString : ''}`;

  try {
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
    return res.json(data);
  } catch (error) {
    console.error('[Submissions] Error fetching history:', error);
    return res.status(502).json({
      error: 'Failed to fetch submission history from backend'
    });
  }
});

/**
 * GET /api/submissions/stats
 *
 * Get current user's submission statistics
 * Requires authentication
 */
router.get('/stats', async (req: Request, res: Response) => {
  const authReq = req as AuthenticatedRequest;

  if (!authReq.user?.userId) {
    return res.status(401).json({
      error: 'Authentication required'
    });
  }

  const userId = authReq.user.userId;
  const url = `${GO_BACKEND_URL}/internal/submissions/user/${userId}/stats`;

  try {
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
    return res.json(data);
  } catch (error) {
    console.error('[Submissions] Error fetching stats:', error);
    return res.status(502).json({
      error: 'Failed to fetch submission stats from backend'
    });
  }
});

/**
 * GET /api/submissions/:id
 *
 * Get submission details by ID
 * Proxied to Go backend
 */
router.get('/:id', submissionsProxy);

export default router;
