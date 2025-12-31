import { Router, Request, Response } from 'express';

const router = Router();

interface SubmissionRequest {
  code: string;
  language: string;
  problemId: number;
}

/**
 * POST /api/submissions
 *
 * Hardcoded submission endpoint for POC
 * Returns fake "Accepted" response with mock metrics
 */
router.post('/', (_req: Request, res: Response): void => {
  const { code, language, problemId } = _req.body as SubmissionRequest;

  // Basic validation
  if (!code || !language || !problemId) {
    res.status(400).json({
      error: 'Missing required fields: code, language, problemId'
    });
    return;
  }

  // Simulate processing delay
  setTimeout(() => {
    // Return hardcoded "Accepted" response with fake metrics
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
  }, 800); // Simulate 800ms processing time
});

/**
 * GET /api/submissions/:id
 *
 * Get submission details (placeholder for future)
 */
router.get('/:id', (_req: Request, res: Response) => {
  return res.json({
    message: 'Submission details endpoint - coming soon'
  });
});

/**
 * GET /api/submissions/history
 *
 * Get user submission history (placeholder for future)
 */
router.get('/history', (_req: Request, res: Response) => {
  return res.json({
    submissions: [],
    message: 'Submission history - coming soon'
  });
});

export default router;
