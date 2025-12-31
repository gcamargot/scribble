import { jest, describe, it, expect, beforeEach } from '@jest/globals';
import request from 'supertest';
import express, { Express } from 'express';

// Set mock mode for testing (avoids actual proxy)
process.env.USE_MOCK_BACKEND = 'true';

// Mock the proxy module to avoid actual HTTP calls
jest.unstable_mockModule('../middleware/proxy.js', () => ({
  submissionsProxy: (_req: unknown, _res: unknown, next: () => void) => next(),
  problemsProxy: (_req: unknown, _res: unknown, next: () => void) => next(),
  leaderboardsProxy: (_req: unknown, _res: unknown, next: () => void) => next(),
  backendProxy: (_req: unknown, _res: unknown, next: () => void) => next()
}));

// Import after mocking
const { default: submissionsRouter } = await import('../routes/submissions.js');

/**
 * Create a test Express app with submissions routes
 */
function createTestApp(): Express {
  const app = express();
  app.use(express.json());
  app.use('/api/submissions', submissionsRouter);
  return app;
}

describe('Submissions Routes', () => {
  let app: Express;

  beforeEach(() => {
    app = createTestApp();
    jest.clearAllMocks();
  });

  describe('POST /api/submissions (Mock Mode)', () => {
    const validSubmission = {
      code: 'function twoSum(nums, target) { return [0, 1]; }',
      language: 'javascript',
      problemId: 1
    };

    it('should return 400 if code is missing', async () => {
      const response = await request(app)
        .post('/api/submissions')
        .send({ language: 'javascript', problemId: 1 });

      expect(response.status).toBe(400);
      expect(response.body.error).toBe('Missing required fields: code, language, problemId');
    });

    it('should return 400 if language is missing', async () => {
      const response = await request(app)
        .post('/api/submissions')
        .send({ code: 'console.log("test")', problemId: 1 });

      expect(response.status).toBe(400);
      expect(response.body.error).toBe('Missing required fields: code, language, problemId');
    });

    it('should return 400 if problemId is missing', async () => {
      const response = await request(app)
        .post('/api/submissions')
        .send({ code: 'console.log("test")', language: 'javascript' });

      expect(response.status).toBe(400);
      expect(response.body.error).toBe('Missing required fields: code, language, problemId');
    });

    it('should return mock success response in mock mode', async () => {
      const response = await request(app)
        .post('/api/submissions')
        .send(validSubmission);

      expect(response.status).toBe(200);
      expect(response.body.success).toBe(true);
      expect(response.body.status).toBe('Accepted');
      expect(response.body.verdict).toBe('Accepted');
      expect(response.body.executionTime).toBe('45ms');
      expect(response.body.memoryUsed).toBe('14.2MB');
      expect(response.body.testsPassed).toBe(3);
      expect(response.body.testsTotal).toBe(3);
      expect(response.body.message).toBe('All test cases passed!');
    }, 10000); // Allow for the 800ms mock delay

    it('should return submission metadata', async () => {
      const response = await request(app)
        .post('/api/submissions')
        .send(validSubmission);

      expect(response.status).toBe(200);
      expect(response.body.submission).toBeDefined();
      expect(response.body.submission.problemId).toBe(validSubmission.problemId);
      expect(response.body.submission.language).toBe(validSubmission.language);
      expect(response.body.submission.codeLength).toBe(validSubmission.code.length);
      expect(response.body.submission.timestamp).toBeDefined();
      expect(response.body.submission.id).toBeDefined();
    }, 10000);

    it('should return percentile metrics', async () => {
      const response = await request(app)
        .post('/api/submissions')
        .send(validSubmission);

      expect(response.status).toBe(200);
      expect(response.body.percentile).toBeDefined();
      expect(response.body.percentile.time).toBe(78);
      expect(response.body.percentile.memory).toBe(65);
    }, 10000);
  });
});
