import rateLimit from 'express-rate-limit';

/**
 * Rate limiter for authentication endpoints
 *
 * Protects against brute force attacks by limiting the number
 * of requests from a single IP address within a time window.
 *
 * Configuration:
 * - 10 requests per 15 minutes per IP
 * - Returns 429 status code when limit exceeded
 * - Uses sliding window to track requests
 */
export const authRateLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 10, // Limit each IP to 10 requests per windowMs
  message: {
    error: 'Too many authentication attempts from this IP, please try again after 15 minutes'
  },
  standardHeaders: true, // Return rate limit info in the `RateLimit-*` headers
  legacyHeaders: false, // Disable the `X-RateLimit-*` headers
  // Skip rate limiting in development if needed
  skip: (_req) => {
    // Only skip in development mode AND if explicitly configured
    return process.env.NODE_ENV === 'development' &&
           process.env.SKIP_RATE_LIMIT === 'true';
  }
});

/**
 * Stricter rate limiter for sensitive operations like password reset
 *
 * Configuration:
 * - 3 requests per 15 minutes per IP
 * - Used for highly sensitive endpoints
 */
export const strictRateLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 3, // Limit each IP to 3 requests per windowMs
  message: {
    error: 'Too many attempts, please try again after 15 minutes'
  },
  standardHeaders: true,
  legacyHeaders: false,
  skip: (_req) => {
    return process.env.NODE_ENV === 'development' &&
           process.env.SKIP_RATE_LIMIT === 'true';
  }
});
