import { createProxyMiddleware, Options } from 'http-proxy-middleware';
import { Request, Response } from 'express';
import { ClientRequest, IncomingMessage, ServerResponse } from 'http';

// Go backend configuration from environment
const GO_BACKEND_URL = process.env.GO_BACKEND_URL || 'http://localhost:8080';
const INTERNAL_AUTH_SECRET = process.env.INTERNAL_AUTH_SECRET || '';

/**
 * Proxy middleware for forwarding requests to Go backend
 *
 * Adds X-Internal-Auth header for service-to-service authentication
 */

// Extend Request type with user context
interface AuthenticatedRequest extends Request {
  user?: {
    userId: string;
    username: string;
  };
}

// Common proxy options
const baseProxyOptions: Options = {
  target: GO_BACKEND_URL,
  changeOrigin: true,
  secure: false, // Allow self-signed certs in development
  timeout: 30000, // 30 second timeout for code execution
  proxyTimeout: 30000,
  onProxyReq: (proxyReq: ClientRequest, req: IncomingMessage) => {
    // Add internal auth header for service-to-service auth
    proxyReq.setHeader('X-Internal-Auth', INTERNAL_AUTH_SECRET);

    // Forward user context if available (from JWT middleware)
    const typedReq = req as AuthenticatedRequest;
    if (typedReq.user) {
      proxyReq.setHeader('X-User-Id', typedReq.user.userId);
      proxyReq.setHeader('X-Username', typedReq.user.username);
    }

    console.log(`[Proxy] ${req.method} ${req.url} -> ${GO_BACKEND_URL}${proxyReq.path}`);
  },
  onProxyRes: (proxyRes: IncomingMessage, req: IncomingMessage) => {
    console.log(`[Proxy] Response from ${req.url}: ${proxyRes.statusCode}`);
  },
  onError: (err: Error, req: IncomingMessage, res: ServerResponse | Response) => {
    console.error(`[Proxy] Error proxying ${req.url}:`, err.message);

    // Type guard for ServerResponse
    const serverRes = res as ServerResponse;
    if (!serverRes.headersSent) {
      serverRes.writeHead(502, { 'Content-Type': 'application/json' });
      serverRes.end(JSON.stringify({
        error: 'Failed to connect to backend service',
        details: process.env.NODE_ENV === 'development' ? err.message : undefined
      }));
    }
  }
};

/**
 * Proxy for code submissions
 * POST /api/submissions -> POST /internal/execute
 */
export const submissionsProxy = createProxyMiddleware({
  ...baseProxyOptions,
  pathRewrite: {
    '^/api/submissions': '/internal/execute'
  }
});

/**
 * Proxy for problems API
 * GET /api/problems/* -> GET /internal/problems/*
 */
export const problemsProxy = createProxyMiddleware({
  ...baseProxyOptions,
  pathRewrite: {
    '^/api/problems': '/internal/problems'
  }
});

/**
 * Proxy for leaderboards API
 * GET /api/leaderboards/* -> GET /internal/leaderboards/*
 */
export const leaderboardsProxy = createProxyMiddleware({
  ...baseProxyOptions,
  pathRewrite: {
    '^/api/leaderboards': '/internal/leaderboards'
  }
});

/**
 * Generic backend proxy (for any /api/backend/* routes)
 * Useful for direct passthrough when needed
 */
export const backendProxy = createProxyMiddleware({
  ...baseProxyOptions,
  pathRewrite: {
    '^/api/backend': '/internal'
  }
});
