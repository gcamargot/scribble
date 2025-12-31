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

/**
 * GET /api/submissions/history
 *
 * Get user submission history
 * Proxied to Go backend
 */
router.get('/history', submissionsProxy);

/**
 * GET /api/submissions/:id
 *
 * Get submission details by ID
 * Proxied to Go backend
 */
router.get('/:id', submissionsProxy);

/**
 * GET /api/submissions/:id/percentile
 *
 * Get percentile comparison metrics for a submission
 * Returns execution time and memory percentile rankings
 * e.g., "Faster than 78% of submissions"
 * Proxied to Go backend
 */
router.get('/:id/percentile', submissionsProxy);

export default router;
