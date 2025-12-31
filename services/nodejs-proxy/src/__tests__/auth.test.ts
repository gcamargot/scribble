import { jest, describe, it, expect, beforeEach } from '@jest/globals';
import request from 'supertest';
import express, { Express } from 'express';
import cookieParser from 'cookie-parser';
import jwt from 'jsonwebtoken';
import axios, { AxiosError } from 'axios';

// Test environment variables - set BEFORE any imports
const TEST_JWT_SECRET = 'test-jwt-secret-for-testing';
const TEST_CLIENT_ID = 'test-client-id';
const TEST_CLIENT_SECRET = 'test-client-secret';
const TEST_REDIRECT_URI = 'http://localhost:3000/api/auth/discord/callback';

process.env.JWT_SECRET = TEST_JWT_SECRET;
process.env.DISCORD_CLIENT_ID = TEST_CLIENT_ID;
process.env.DISCORD_CLIENT_SECRET = TEST_CLIENT_SECRET;
process.env.DISCORD_REDIRECT_URI = TEST_REDIRECT_URI;

// Mock modules
jest.unstable_mockModule('axios', () => ({
  default: {
    post: jest.fn(),
    get: jest.fn(),
    isAxiosError: (error: unknown): error is AxiosError =>
      error !== null && typeof error === 'object' && 'isAxiosError' in error,
  },
  isAxiosError: (error: unknown): error is AxiosError =>
    error !== null && typeof error === 'object' && 'isAxiosError' in error,
}));

jest.unstable_mockModule('../middleware/rateLimiter.js', () => ({
  authRateLimiter: (_req: unknown, _res: unknown, next: () => void) => next(),
}));

// Import modules after mocks are set up
const mockedAxios = (await import('axios')).default as jest.Mocked<typeof axios>;
const { default: authRouter } = await import('../routes/auth.js');

// Helper to create axios error for testing
function createAxiosError(status: number, data: unknown): AxiosError {
  const error = new Error('Axios error') as AxiosError;
  error.isAxiosError = true;
  error.response = {
    status,
    data,
    statusText: 'Error',
    headers: {},
    config: {} as never,
  };
  return error;
}

/**
 * Create a test Express app with auth routes
 */
function createTestApp(): Express {
  const app = express();
  app.use(express.json());
  app.use(cookieParser());
  app.use('/api/auth', authRouter);
  return app;
}

describe('Auth Routes', () => {
  let app: Express;

  beforeEach(() => {
    app = createTestApp();
    jest.clearAllMocks();
  });

  describe('POST /api/auth/discord/callback', () => {
    const mockDiscordUser = {
      id: '123456789',
      username: 'testuser',
      discriminator: '1234',
      avatar: 'abc123',
    };

    const mockTokenResponse = {
      access_token: 'mock-access-token',
      token_type: 'Bearer',
      expires_in: 604800,
      refresh_token: 'mock-refresh-token',
      scope: 'identify guilds',
    };

    it('should return 400 if authorization code is missing', async () => {
      const response = await request(app)
        .post('/api/auth/discord/callback')
        .send({});

      expect(response.status).toBe(400);
      expect(response.body.error).toBe('Authorization code is required');
    });

    it('should return 400 if code is empty string', async () => {
      const response = await request(app)
        .post('/api/auth/discord/callback')
        .send({ code: '' });

      expect(response.status).toBe(400);
      expect(response.body.error).toBe('Authorization code is required');
    });

    it('should successfully authenticate and return user data', async () => {
      // Mock Discord token exchange
      mockedAxios.post.mockResolvedValueOnce({ data: mockTokenResponse });
      // Mock Discord user fetch
      mockedAxios.get.mockResolvedValueOnce({ data: mockDiscordUser });

      const response = await request(app)
        .post('/api/auth/discord/callback')
        .send({ code: 'valid-auth-code' });

      expect(response.status).toBe(200);
      expect(response.body.success).toBe(true);
      expect(response.body.user).toEqual({
        id: mockDiscordUser.id,
        username: mockDiscordUser.username,
        discriminator: mockDiscordUser.discriminator,
        avatar: `https://cdn.discordapp.com/avatars/${mockDiscordUser.id}/${mockDiscordUser.avatar}.png`,
      });

      // Check that a cookie was set
      expect(response.headers['set-cookie']).toBeDefined();
      expect(response.headers['set-cookie'][0]).toMatch(/scribble_auth=/);
    });

    it('should return null avatar if user has no avatar', async () => {
      const userWithoutAvatar = { ...mockDiscordUser, avatar: null };
      mockedAxios.post.mockResolvedValueOnce({ data: mockTokenResponse });
      mockedAxios.get.mockResolvedValueOnce({ data: userWithoutAvatar });

      const response = await request(app)
        .post('/api/auth/discord/callback')
        .send({ code: 'valid-auth-code' });

      expect(response.status).toBe(200);
      expect(response.body.user.avatar).toBeNull();
    });

    it('should NOT return accessToken in response (security)', async () => {
      mockedAxios.post.mockResolvedValueOnce({ data: mockTokenResponse });
      mockedAxios.get.mockResolvedValueOnce({ data: mockDiscordUser });

      const response = await request(app)
        .post('/api/auth/discord/callback')
        .send({ code: 'valid-auth-code' });

      expect(response.status).toBe(200);
      // Verify accessToken is NOT in response
      expect(response.body.accessToken).toBeUndefined();
      expect(response.body.user.accessToken).toBeUndefined();
    });

    it('should handle Discord API token exchange error', async () => {
      const axiosError = createAxiosError(401, { error: 'invalid_grant' });
      mockedAxios.post.mockRejectedValueOnce(axiosError);

      const response = await request(app)
        .post('/api/auth/discord/callback')
        .send({ code: 'invalid-auth-code' });

      expect(response.status).toBe(401);
      expect(response.body.error).toBe('Failed to authenticate with Discord');
    });

    it('should handle Discord API user fetch error', async () => {
      mockedAxios.post.mockResolvedValueOnce({ data: mockTokenResponse });
      const axiosError = createAxiosError(403, { message: 'Forbidden' });
      mockedAxios.get.mockRejectedValueOnce(axiosError);

      const response = await request(app)
        .post('/api/auth/discord/callback')
        .send({ code: 'valid-auth-code' });

      expect(response.status).toBe(403);
      expect(response.body.error).toBe('Failed to authenticate with Discord');
    });

    it('should handle non-Axios errors', async () => {
      const nonAxiosError = new Error('Network error');
      mockedAxios.post.mockRejectedValueOnce(nonAxiosError);

      const response = await request(app)
        .post('/api/auth/discord/callback')
        .send({ code: 'valid-auth-code' });

      expect(response.status).toBe(500);
      expect(response.body.error).toBe('Internal server error during authentication');
    });
  });

  describe('POST /api/auth/logout', () => {
    it('should successfully logout and clear cookie', async () => {
      const response = await request(app)
        .post('/api/auth/logout');

      expect(response.status).toBe(200);
      expect(response.body.success).toBe(true);
      expect(response.body.message).toBe('Logged out successfully');

      // Check that cookie is cleared
      expect(response.headers['set-cookie']).toBeDefined();
      expect(response.headers['set-cookie'][0]).toMatch(/scribble_auth=;/);
    });
  });

  describe('GET /api/auth/me', () => {
    const userPayload = {
      userId: '123456789',
      username: 'testuser',
      discriminator: '1234',
      avatar: 'abc123',
    };

    it('should return 401 if no token is provided', async () => {
      const response = await request(app)
        .get('/api/auth/me');

      expect(response.status).toBe(401);
      expect(response.body.error).toBe('Not authenticated');
    });

    it('should return 401 for invalid token', async () => {
      const response = await request(app)
        .get('/api/auth/me')
        .set('Cookie', 'scribble_auth=invalid-token');

      expect(response.status).toBe(401);
      expect(response.body.error).toBe('Invalid or expired token');
    });

    it('should return 401 for expired token', async () => {
      const expiredToken = jwt.sign(userPayload, TEST_JWT_SECRET, {
        expiresIn: '-1s', // Already expired
      });

      const response = await request(app)
        .get('/api/auth/me')
        .set('Cookie', `scribble_auth=${expiredToken}`);

      expect(response.status).toBe(401);
      expect(response.body.error).toBe('Invalid or expired token');
    });

    it('should return user data for valid token', async () => {
      const validToken = jwt.sign(userPayload, TEST_JWT_SECRET, {
        expiresIn: '7d',
      });

      const response = await request(app)
        .get('/api/auth/me')
        .set('Cookie', `scribble_auth=${validToken}`);

      expect(response.status).toBe(200);
      expect(response.body.user).toBeDefined();
      expect(response.body.user.userId).toBe(userPayload.userId);
      expect(response.body.user.username).toBe(userPayload.username);
    });

    it('should return 401 for token signed with wrong secret', async () => {
      const wrongSecretToken = jwt.sign(userPayload, 'wrong-secret', {
        expiresIn: '7d',
      });

      const response = await request(app)
        .get('/api/auth/me')
        .set('Cookie', `scribble_auth=${wrongSecretToken}`);

      expect(response.status).toBe(401);
      expect(response.body.error).toBe('Invalid or expired token');
    });
  });
});
